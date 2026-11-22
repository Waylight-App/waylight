package domain

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"waylight/internal/shared/kernel"
)

// Errors of the module.
var (
	ErrNotConfigured = kernel.NewPublicError(kernel.ErrNotFound, "Grafana is not configured yet.")
	ErrUnreachable   = kernel.NewPublicError(kernel.ErrUnavailable, "Grafana is not reachable.")
	ErrNoRenderer    = kernel.NewPublicError(kernel.ErrUnavailable, "Grafana cannot render this panel as an image; install the Grafana image renderer.")
	ErrBadDashboard  = kernel.NewPublicError(kernel.ErrNotFound, "The dashboard does not exist.")
	ErrNoData        = kernel.NewPublicError(kernel.ErrUnavailable, "Grafana returned no values for this panel; panels that use dashboard variables need the image renderer.")
)

// Settings are the connection details of Grafana.
type Settings struct {
	URL   string
	Token string
}

// NewSettings validates and normalises the settings.
func NewSettings(rawURL, token string) (Settings, error) {
	rawURL = strings.TrimRight(strings.TrimSpace(rawURL), "/")
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return Settings{}, kernel.NewValidationError("url", "The address must look like http://192.168.1.2:3000.")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return Settings{}, kernel.NewValidationError("token", "The service account token is required.")
	}
	return Settings{URL: rawURL, Token: token}, nil
}

// uidPattern matches Grafana dashboard UIDs (up to 40 letters, digits, - and _).
var uidPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,40}$`)

// ValidUID reports whether uid is a dashboard UID.
func ValidUID(uid string) bool { return uidPattern.MatchString(uid) }

// Dashboard is a dashboard found by the search.
type Dashboard struct {
	UID, Title, Folder string
	// URL is the dashboard's address in Grafana.
	URL string
}

// Panel is a panel of a dashboard.
type Panel struct {
	ID          int
	Title, Type string
}

// AlertState is the state of an alert instance.
type AlertState string

// The states shown; normal alerts are left out.
const (
	Firing  AlertState = "firing"
	Pending AlertState = "pending"
)

// ParseAlertState maps the states of Grafana's alert API ("Alerting",
// "firing", "Pending") to AlertState; false for normal or unknown ones.
func ParseAlertState(s string) (AlertState, bool) {
	switch strings.ToLower(s) {
	case "alerting", "firing":
		return Firing, true
	case "pending":
		return Pending, true
	default:
		return "", false
	}
}

// Alert is a firing or pending alert instance.
type Alert struct {
	Name, Summary, Severity string
	State                   AlertState
	Since                   time.Time
}

// Image is a rendered panel.
type Image struct {
	ContentType string
	Data        []byte
}

// RenderRequest describes a panel image.
type RenderRequest struct {
	UID     string
	PanelID int
	Width   int
	Height  int
	// Range is a Grafana relative range such as "6h" (now-6h to now).
	Range string
	Dark  bool
}

var rangePattern = regexp.MustCompile(`^[1-9][0-9]{0,2}[mhdw]$`)

// NewRenderRequest validates the request and fills defaults.
func NewRenderRequest(uid string, panelID, width, height int, rng string, dark bool) (RenderRequest, error) {
	if !ValidUID(uid) || panelID <= 0 {
		return RenderRequest{}, ErrBadDashboard
	}
	if rng == "" {
		rng = "6h"
	}
	if !rangePattern.MatchString(rng) {
		return RenderRequest{}, kernel.NewValidationError("range", "The time range must look like 30m, 6h, 7d or 4w.")
	}
	return RenderRequest{UID: uid, PanelID: panelID, Width: clamp(width, 200, 1600, 600), Height: clamp(height, 100, 1200, 300),
		Range: rng, Dark: dark}, nil
}

func clamp(v, lo, hi, fallback int) int {
	if v == 0 {
		return fallback
	}
	return max(lo, min(hi, v))
}

// Limits of the values read without the image renderer (#579): enough for a
// stat and a sparkline in a widget, small enough for a phone.
const (
	MaxSeries = 5
	MaxPoints = 120
)

// Point is one value of a series.
type Point struct {
	Time  time.Time
	Value float64
}

// Series is one line of a panel, oldest point first.
type Series struct {
	Name   string
	Points []Point
}

// PanelData are the values of a panel, read through Grafana's query API
// instead of the image renderer (#579).
type PanelData struct {
	Title string
	// Unit is Grafana's unit id of the panel, e.g. "percent" or "celsius".
	Unit string
	// Decimals of the panel; nil lets the client choose.
	Decimals *int
	Series   []Series
}

// DataRequest describes the values of a panel.
type DataRequest struct {
	UID     string
	PanelID int
	// Range is a Grafana relative range such as "6h" (now-6h to now).
	Range string
}

// NewDataRequest validates the request; an empty range means six hours.
func NewDataRequest(uid string, panelID int, rng string) (DataRequest, error) {
	r, err := NewRenderRequest(uid, panelID, 0, 0, rng, false)
	if err != nil {
		return DataRequest{}, err
	}
	return DataRequest{UID: r.UID, PanelID: r.PanelID, Range: r.Range}, nil
}

// Downsample keeps at most max points of a series: the newest point always
// stays (it is the value a stat shows), the others are picked evenly.
func Downsample(points []Point, maxPoints int) []Point {
	if len(points) <= maxPoints || maxPoints < 2 {
		return points
	}
	out := make([]Point, 0, maxPoints)
	step := float64(len(points)-1) / float64(maxPoints-1)
	for i := range maxPoints {
		out = append(out, points[int(float64(i)*step+0.5)])
	}
	out[len(out)-1] = points[len(points)-1]
	return out
}

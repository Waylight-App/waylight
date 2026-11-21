// Package presenter maps Grafana use case results to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/grafana/domain"
	"waylight/internal/modules/grafana/usecase"
)

// SettingsDTO is returned by GET/PUT /api/grafana/settings; the token is
// never returned.
type SettingsDTO struct {
	Configured bool   `json:"configured"`
	URL        string `json:"url"`
	HasToken   bool   `json:"hasToken"`
}

// AlertDTO is a firing or pending alert.
type AlertDTO struct {
	Name     string     `json:"name"`
	Summary  string     `json:"summary"`
	Severity string     `json:"severity"`
	State    string     `json:"state"`
	Since    *time.Time `json:"since,omitempty"`
}

// AlertsDTO is returned by GET /api/grafana/alerts.
type AlertsDTO struct {
	Alerts []AlertDTO `json:"alerts"`
	// URL opens the alert list in Grafana.
	URL string `json:"url"`
}

// DashboardDTO is a dashboard found by the search.
type DashboardDTO struct {
	UID    string `json:"uid"`
	Title  string `json:"title"`
	Folder string `json:"folder"`
	URL    string `json:"url"`
}

// PanelDTO is a panel of a dashboard.
type PanelDTO struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// Settings maps a SettingsView.
func Settings(v usecase.SettingsView) SettingsDTO { return SettingsDTO(v) }

// Alerts maps the alerts; baseURL links to Grafana's alert list.
func Alerts(list []domain.Alert, baseURL string) AlertsDTO {
	out := AlertsDTO{Alerts: make([]AlertDTO, 0, len(list))}
	if baseURL != "" {
		out.URL = baseURL + "/alerting/list"
	}
	for _, a := range list {
		d := AlertDTO{Name: a.Name, Summary: a.Summary, Severity: a.Severity, State: string(a.State)}
		if !a.Since.IsZero() {
			since := a.Since.UTC()
			d.Since = &since
		}
		out.Alerts = append(out.Alerts, d)
	}
	return out
}

// Dashboards maps the search result.
func Dashboards(list []domain.Dashboard) []DashboardDTO {
	out := make([]DashboardDTO, len(list))
	for i, d := range list {
		out[i] = DashboardDTO{UID: d.UID, Title: d.Title, Folder: d.Folder, URL: d.URL}
	}
	return out
}

// Panels maps the panels.
func Panels(list []domain.Panel) []PanelDTO {
	out := make([]PanelDTO, len(list))
	for i, p := range list {
		out[i] = PanelDTO(p)
	}
	return out
}

// PointDTO is one value; t is the time in Unix milliseconds, 0 when the
// value has no time (a reduced value).
type PointDTO struct {
	T int64   `json:"t"`
	V float64 `json:"v"`
}

// SeriesDTO is one line of a panel, oldest point first.
type SeriesDTO struct {
	Name   string     `json:"name"`
	Points []PointDTO `json:"points"`
}

// PanelDataDTO is returned by GET /api/grafana/data/{uid}/{panel} (#579).
type PanelDataDTO struct {
	Title string `json:"title"`
	// Unit is Grafana's unit id, e.g. "percent"; empty when the panel has none.
	Unit     string      `json:"unit"`
	Decimals *int        `json:"decimals,omitempty"`
	Series   []SeriesDTO `json:"series"`
}

// PanelData maps the values of a panel.
func PanelData(d domain.PanelData) PanelDataDTO {
	out := PanelDataDTO{Title: d.Title, Unit: d.Unit, Decimals: d.Decimals, Series: make([]SeriesDTO, 0, len(d.Series))}
	for _, s := range d.Series {
		points := make([]PointDTO, len(s.Points))
		for i, p := range s.Points {
			if !p.Time.IsZero() {
				points[i].T = p.Time.UnixMilli()
			}
			points[i].V = p.Value
		}
		out.Series = append(out.Series, SeriesDTO{Name: s.Name, Points: points})
	}
	return out
}

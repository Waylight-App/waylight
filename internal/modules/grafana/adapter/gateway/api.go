package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"waylight/internal/modules/grafana/domain"
	"waylight/internal/shared/kernel"
)

const (
	maxJSONBytes  = 4 << 20
	maxImageBytes = 4 << 20
	// maxDashboards bounds a search answer.
	maxDashboards = 50
)

// API implements usecase.Client with a service account token (Viewer role).
type API struct{ client *http.Client }

// NewAPI returns a client.
func NewAPI(client *http.Client) *API { return &API{client: client} }

// Alerts implements usecase.Client through the Prometheus-compatible alert
// list of Grafana Alerting.
func (a *API) Alerts(ctx context.Context, s domain.Settings) ([]domain.Alert, error) {
	var payload struct {
		Data struct {
			Alerts []struct {
				Labels      map[string]string `json:"labels"`
				Annotations map[string]string `json:"annotations"`
				State       string            `json:"state"`
				ActiveAt    string            `json:"activeAt"`
			} `json:"alerts"`
		} `json:"data"`
	}
	if err := a.getJSON(ctx, s, "/api/prometheus/grafana/api/v1/alerts", &payload); err != nil {
		return nil, err
	}
	out := []domain.Alert{}
	for _, al := range payload.Data.Alerts {
		state, ok := domain.ParseAlertState(al.State)
		if !ok {
			continue
		}
		since, _ := time.Parse(time.RFC3339Nano, al.ActiveAt)
		summary := al.Annotations["summary"]
		if summary == "" {
			summary = al.Annotations["description"]
		}
		out = append(out, domain.Alert{Name: al.Labels["alertname"], Summary: summary, Severity: al.Labels["severity"], State: state, Since: since})
	}
	return out, nil
}

// Search implements usecase.Client: dashboards whose title matches query.
func (a *API) Search(ctx context.Context, s domain.Settings, query string) ([]domain.Dashboard, error) {
	var payload []struct {
		UID         string `json:"uid"`
		Title       string `json:"title"`
		URL         string `json:"url"`
		FolderTitle string `json:"folderTitle"`
	}
	q := url.Values{"type": {"dash-db"}, "limit": {strconv.Itoa(maxDashboards)}, "query": {query}}
	if err := a.getJSON(ctx, s, "/api/search?"+q.Encode(), &payload); err != nil {
		return nil, err
	}
	out := make([]domain.Dashboard, 0, len(payload))
	for _, d := range payload {
		if domain.ValidUID(d.UID) {
			out = append(out, domain.Dashboard{UID: d.UID, Title: d.Title, Folder: d.FolderTitle, URL: s.URL + d.URL})
		}
	}
	return out, nil
}

type panelJSON struct {
	ID     int         `json:"id"`
	Title  string      `json:"title"`
	Type   string      `json:"type"`
	Panels []panelJSON `json:"panels"`
	// The fields below are only read for the values of a panel (#579).
	Datasource  json.RawMessage              `json:"datasource"`
	Targets     []map[string]json.RawMessage `json:"targets"`
	FieldConfig struct {
		Defaults struct {
			Unit     string `json:"unit"`
			Decimals *int   `json:"decimals"`
		} `json:"defaults"`
	} `json:"fieldConfig"`
}

// Panels implements usecase.Client: the panels of a dashboard, panels in
// collapsed rows included; rows themselves are left out.
func (a *API) Panels(ctx context.Context, s domain.Settings, uid string) ([]domain.Panel, error) {
	var payload struct {
		Dashboard struct {
			Panels []panelJSON `json:"panels"`
		} `json:"dashboard"`
	}
	if err := a.getJSON(ctx, s, "/api/dashboards/uid/"+url.PathEscape(uid), &payload); err != nil {
		return nil, err
	}
	var out []domain.Panel
	var walk func([]panelJSON)
	walk = func(list []panelJSON) {
		for _, p := range list {
			if p.Type == "row" {
				walk(p.Panels)
				continue
			}
			if p.ID > 0 {
				out = append(out, domain.Panel{ID: p.ID, Title: p.Title, Type: p.Type})
			}
		}
	}
	walk(payload.Dashboard.Panels)
	if out == nil {
		out = []domain.Panel{}
	}
	return out, nil
}

// Render implements usecase.Client with the image renderer (/render/d-solo).
// Without the renderer Grafana answers with an error or an HTML page, which
// becomes domain.ErrNoRenderer.
func (a *API) Render(ctx context.Context, s domain.Settings, r domain.RenderRequest) (domain.Image, error) {
	theme := "light"
	if r.Dark {
		theme = "dark"
	}
	q := url.Values{
		"panelId": {strconv.Itoa(r.PanelID)}, "width": {strconv.Itoa(r.Width)}, "height": {strconv.Itoa(r.Height)},
		"from": {"now-" + r.Range}, "to": {"now"}, "theme": {theme},
	}
	res, err := a.get(ctx, s, "/render/d-solo/"+url.PathEscape(r.UID)+"/_?"+q.Encode(), "image/png")
	if err != nil {
		return domain.Image{}, domain.ErrNoRenderer
	}
	defer res.Body.Close()
	ct, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if ct != "image/png" {
		return domain.Image{}, domain.ErrNoRenderer
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxImageBytes+1))
	if err != nil || len(data) > maxImageBytes {
		return domain.Image{}, domain.ErrNoRenderer
	}
	return domain.Image{ContentType: ct, Data: data}, nil
}

func (a *API) getJSON(ctx context.Context, s domain.Settings, path string, dst any) error {
	res, err := a.get(ctx, s, path, "application/json")
	if err != nil {
		return err
	}
	return decode(res, dst)
}

func decode(res *http.Response, dst any) error {
	defer res.Body.Close()
	if err := json.NewDecoder(io.LimitReader(res.Body, maxJSONBytes)).Decode(dst); err != nil {
		return fmt.Errorf("%w: %w", kernel.ErrTargetWrongService, err)
	}
	return nil
}

func (a *API) get(ctx context.Context, s domain.Settings, path, accept string) (*http.Response, error) {
	return a.do(ctx, s, http.MethodGet, path, accept, nil)
}

func (a *API) do(ctx context.Context, s domain.Settings, method, path, accept string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.URL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("Accept", accept)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	switch res.StatusCode {
	case http.StatusOK:
		return res, nil
	case http.StatusUnauthorized:
		_ = res.Body.Close()
		return nil, fmt.Errorf("%w: the token was rejected (%s)", kernel.ErrTargetAuth, res.Status)
	case http.StatusForbidden:
		_ = res.Body.Close()
		return nil, fmt.Errorf("%w: the token lacks a permission (%s)", kernel.ErrTargetAuth, res.Status)
	case http.StatusNotFound:
		_ = res.Body.Close()
		return nil, domain.ErrBadDashboard
	default:
		_ = res.Body.Close()
		return nil, fmt.Errorf("unexpected answer from Grafana: %s", res.Status)
	}
}

package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"waylight/internal/modules/grafana/domain"
)

func fakeGrafana(t *testing.T, renderer bool) (*httptest.Server, domain.Settings) {
	t.Helper()
	mux := http.NewServeMux()
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer glsa_ok" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("GET /api/prometheus/grafana/api/v1/alerts", auth(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"success","data":{"alerts":[
			{"labels":{"alertname":"Disk full","severity":"critical"},"annotations":{"summary":"/ at 95 %"},"state":"Alerting","activeAt":"2026-11-20T08:00:00Z"},
			{"labels":{"alertname":"CPU high"},"annotations":{"description":"load 8"},"state":"Pending","activeAt":"2026-11-20T09:00:00Z"},
			{"labels":{"alertname":"Fine"},"state":"Normal"}]}}`))
	}))
	mux.HandleFunc("GET /api/search", auth(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("type") != "dash-db" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`[{"uid":"node","title":"Node Exporter","url":"/d/node/node-exporter","folderTitle":"Servers"},{"uid":"../bad","title":"x","url":"/x"}]`))
	}))
	mux.HandleFunc("GET /api/dashboards/uid/node", auth(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"dashboard":{"panels":[{"id":1,"title":"CPU","type":"timeseries"},
			{"id":2,"type":"row","title":"More","panels":[{"id":3,"title":"Memory","type":"gauge"}]}]}}`))
	}))
	mux.HandleFunc("GET /render/d-solo/node/_", auth(func(w http.ResponseWriter, r *http.Request) {
		if !renderer {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>no renderer</html>"))
			return
		}
		if r.URL.Query().Get("from") != "now-6h" || r.URL.Query().Get("theme") != "dark" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("\x89PNG"))
	}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, domain.Settings{URL: srv.URL, Token: "glsa_ok"}
}

func TestAPI(t *testing.T) {
	_, s := fakeGrafana(t, true)
	api := NewAPI(http.DefaultClient)
	ctx := context.Background()

	alerts, err := api.Alerts(ctx, s)
	if err != nil || len(alerts) != 2 || alerts[0].State != domain.Firing || alerts[0].Summary != "/ at 95 %" ||
		alerts[0].Severity != "critical" || alerts[1].Summary != "load 8" || alerts[1].Since.Hour() != 9 {
		t.Fatalf("alerts = %+v, %v", alerts, err)
	}
	dashboards, err := api.Search(ctx, s, "node")
	if err != nil || len(dashboards) != 1 || dashboards[0].URL != s.URL+"/d/node/node-exporter" || dashboards[0].Folder != "Servers" {
		t.Fatalf("search = %+v, %v", dashboards, err)
	}
	panels, err := api.Panels(ctx, s, "node")
	if err != nil || len(panels) != 2 || panels[1].Title != "Memory" {
		t.Fatalf("panels = %+v, %v", panels, err)
	}
	r, _ := domain.NewRenderRequest("node", 1, 0, 0, "6h", true)
	img, err := api.Render(ctx, s, r)
	if err != nil || img.ContentType != "image/png" {
		t.Fatalf("render = %+v, %v", img, err)
	}
	if _, err := api.Search(ctx, domain.Settings{URL: s.URL, Token: "wrong"}, ""); err == nil {
		t.Fatal("a wrong token must fail")
	}
	if _, err := api.Panels(ctx, s, "missing"); !errors.Is(err, domain.ErrBadDashboard) {
		t.Fatalf("missing dashboard: %v", err)
	}
}

func TestRenderWithoutRenderer(t *testing.T) {
	_, s := fakeGrafana(t, false)
	r, _ := domain.NewRenderRequest("node", 1, 0, 0, "6h", true)
	if _, err := NewAPI(http.DefaultClient).Render(context.Background(), s, r); !errors.Is(err, domain.ErrNoRenderer) {
		t.Fatalf("err = %v", err)
	}
}

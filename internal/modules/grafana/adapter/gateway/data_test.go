package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"waylight/internal/modules/grafana/domain"
)

// dashboardWithTargets has a panel with a hidden target, a target with its
// own data source, one that inherits the panel's and one on the "mixed"
// pseudo source, which cannot be queried.
const dashboardWithTargets = `{"dashboard":{"panels":[
	{"id":2,"type":"row","title":"Host","panels":[
		{"id":7,"title":"CPU","type":"timeseries",
		 "datasource":{"type":"prometheus","uid":"prom"},
		 "fieldConfig":{"defaults":{"unit":"percent","decimals":1}},
		 "targets":[
			{"refId":"A","expr":"cpu_busy"},
			{"refId":"B","expr":"hidden","hide":true},
			{"refId":"C","expr":"load","datasource":{"type":"influxdb","uid":"influx"}},
			{"refId":"D","expr":"x","datasource":{"uid":"-- Mixed --"}}
		 ]}
	]},
	{"id":8,"title":"Text","type":"text"}
]}}`

const queryAnswer = `{"results":{
	"C":{"frames":[{"schema":{"name":"load","fields":[{"name":"Time","type":"time"},{"name":"Value","type":"number"}]},
		"data":{"values":[[1795262400000,1795262460000],[0.5,0.7]]}}]},
	"A":{"frames":[{"schema":{"fields":[
			{"name":"Time","type":"time"},
			{"name":"instance","type":"string"},
			{"name":"Value","type":"number","labels":{"instance":"pi","job":"node"}}]},
		"data":{"values":[[1795262400000,1795262460000,1795262520000],["pi","pi","pi"],[12.5,null,14]]}}]},
	"D":{"error":"not queried"}
}}`

func TestPanelData(t *testing.T) {
	var sent struct {
		From    string           `json:"from"`
		To      string           `json:"to"`
		Queries []map[string]any `json:"queries"`
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/dashboards/uid/node", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(dashboardWithTargets))
	})
	mux.HandleFunc("POST /api/ds/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer glsa_ok" || r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(queryAnswer))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	s := domain.Settings{URL: srv.URL, Token: "glsa_ok"}
	api := NewAPI(srv.Client())

	d, err := api.PanelData(context.Background(), s, domain.DataRequest{UID: "node", PanelID: 7, Range: "6h"})
	if err != nil {
		t.Fatal(err)
	}
	if sent.From != "now-6h" || sent.To != "now" || len(sent.Queries) != 2 {
		t.Fatalf("request: %+v", sent)
	}
	if ds := sent.Queries[0]["datasource"].(map[string]any); ds["uid"] != "prom" {
		t.Errorf("A inherits the panel's data source, got %v", ds)
	}
	if ds := sent.Queries[1]["datasource"].(map[string]any); ds["uid"] != "influx" {
		t.Errorf("C keeps its own data source, got %v", ds)
	}
	if sent.Queries[0]["maxDataPoints"].(float64) != domain.MaxPoints || sent.Queries[0]["intervalMs"].(float64) != 180000 {
		t.Errorf("interval: %v", sent.Queries[0])
	}
	if d.Title != "CPU" || d.Unit != "percent" || d.Decimals == nil || *d.Decimals != 1 {
		t.Errorf("panel: %+v", d)
	}
	if len(d.Series) != 2 || d.Series[0].Name != "instance=pi, job=node" || d.Series[1].Name != "load" {
		t.Fatalf("series: %+v", d.Series)
	}
	// The null is skipped and the string column ignored.
	if p := d.Series[0].Points; len(p) != 2 || p[1].Value != 14 || p[1].Time.UnixMilli() != 1795262520000 {
		t.Errorf("points: %+v", p)
	}

	if _, err := api.PanelData(context.Background(), s, domain.DataRequest{UID: "node", PanelID: 99, Range: "6h"}); !errors.Is(err, domain.ErrBadDashboard) {
		t.Errorf("missing panel: %v", err)
	}
	if _, err := api.PanelData(context.Background(), s, domain.DataRequest{UID: "node", PanelID: 8, Range: "6h"}); !errors.Is(err, domain.ErrNoData) {
		t.Errorf("panel without queries: %v", err)
	}
}

func TestSeriesNameOrder(t *testing.T) {
	if got := seriesName("", "", nil, "Value", "", "A"); got != "A" {
		t.Errorf("fallback to the query: %q", got)
	}
	if got := seriesName("", "Shown", map[string]string{"a": "b"}, "x", "f", "A"); got != "Shown" {
		t.Errorf("display name first: %q", got)
	}
	if got := seriesName("", "", nil, "temp", "f", "A"); got != "temp" {
		t.Errorf("field name: %q", got)
	}
}

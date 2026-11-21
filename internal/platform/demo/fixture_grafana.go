package demo

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"time"
)

// GrafanaURL is the example Grafana of the showcase (#579). It has no image
// renderer, so the panel widget shows the values it reads through the query
// API, which is what most home installations see.
const GrafanaURL = "http://grafana.demo.invalid:3000"

// GrafanaDashboard is the UID of its example dashboard; panel 1 is the CPU.
const GrafanaDashboard = "homeserver"

func init() { handle(hostIs("grafana.demo.invalid"), grafana) }

const grafanaDashboardJSON = `{"dashboard":{"uid":"homeserver","title":"Home server","panels":[
	{"id":1,"title":"CPU","type":"timeseries","datasource":{"type":"prometheus","uid":"prom"},
	 "fieldConfig":{"defaults":{"unit":"percent","decimals":0}},"targets":[{"refId":"A","expr":"cpu"}]},
	{"id":2,"title":"Temperature","type":"stat","datasource":{"type":"prometheus","uid":"prom"},
	 "fieldConfig":{"defaults":{"unit":"celsius","decimals":1}},"targets":[{"refId":"A","expr":"temperature"}]},
	{"id":3,"title":"Network","type":"timeseries","datasource":{"type":"prometheus","uid":"prom"},
	 "fieldConfig":{"defaults":{"unit":"bps"}},"targets":[{"refId":"A","expr":"rx","legendFormat":"in"},{"refId":"B","expr":"tx","legendFormat":"out"}]}
]}}`

func grafana(req *http.Request, now time.Time) (int, string, string) {
	switch req.URL.Path {
	case "/api/search":
		return jsonOK(`[{"uid":"homeserver","title":"Home server","url":"/d/homeserver/home-server","folderTitle":"Servers"}]`)
	case "/api/dashboards/uid/homeserver":
		return jsonOK(grafanaDashboardJSON)
	case "/api/ds/query":
		return jsonOK(grafanaQuery(req, now))
	case "/api/prometheus/grafana/api/v1/alerts":
		ago := func(m int) string { return now.Add(-time.Duration(m) * time.Minute).UTC().Format(time.RFC3339) }
		return jsonOK(`{"status":"success","data":{"alerts":[
			{"labels":{"alertname":"Backup disk almost full","severity":"warning"},"annotations":{"summary":"/mnt/backup at 91 %"},"state":"Alerting","activeAt":"` + ago(95) + `"},
			{"labels":{"alertname":"Raspberry Pi warm"},"annotations":{"summary":"CPU at 68 °C"},"state":"Pending","activeAt":"` + ago(4) + `"}]}}`)
	}
	// No image renderer: Grafana answers /render with its HTML page.
	return http.StatusOK, "text/html", "<html><body>Grafana</body></html>"
}

// grafanaQuery answers /api/ds/query with one frame per query, a gentle
// curve over the last six hours that ends at now.
func grafanaQuery(req *http.Request, now time.Time) string {
	var body struct {
		Queries []struct {
			RefID string `json:"refId"`
			Expr  string `json:"expr"`
		} `json:"queries"`
	}
	if req.Body != nil {
		_ = json.NewDecoder(io.LimitReader(req.Body, 64<<10)).Decode(&body)
	}
	const points = 72
	results := map[string]any{}
	for _, q := range body.Queries {
		base, swing, name := 0.0, 0.0, q.Expr
		switch q.Expr {
		case "cpu":
			base, swing = 18, 12
		case "temperature":
			base, swing = 52, 6
		case "rx":
			base, swing, name = 4.2e6, 3e6, "in"
		case "tx":
			base, swing, name = 0.9e6, 0.6e6, "out"
		}
		times, values := make([]int64, points), make([]float64, points)
		for i := range points {
			t := now.Add(-time.Duration(points-1-i) * 5 * time.Minute)
			times[i] = t.UnixMilli()
			values[i] = math.Round((base+swing*math.Sin(float64(i)/9)*math.Cos(float64(i)/23))*10) / 10
		}
		results[q.RefID] = map[string]any{"frames": []any{map[string]any{
			"schema": map[string]any{"fields": []any{
				map[string]any{"name": "Time", "type": "time"},
				map[string]any{"name": "Value", "type": "number", "config": map[string]any{"displayNameFromDS": name}},
			}},
			"data": map[string]any{"values": []any{times, values}},
		}}}
	}
	return mustJSON(map[string]any{"results": results})
}

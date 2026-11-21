package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"waylight/internal/modules/grafana/domain"
)

// PanelData implements usecase.Client without the image renderer (#579): it
// reads the panel's queries from the dashboard and runs them through
// /api/ds/query, which a Viewer service account may call. Dashboard
// variables are not resolved; such queries usually return nothing.
func (a *API) PanelData(ctx context.Context, s domain.Settings, r domain.DataRequest) (domain.PanelData, error) {
	panel, err := a.panel(ctx, s, r.UID, r.PanelID)
	if err != nil {
		return domain.PanelData{}, err
	}
	queries := panelQueries(panel, r.Range)
	if len(queries) == 0 {
		return domain.PanelData{}, domain.ErrNoData
	}
	body, err := json.Marshal(map[string]any{"queries": queries, "from": "now-" + r.Range, "to": "now"})
	if err != nil {
		return domain.PanelData{}, fmt.Errorf("encode query: %w", err)
	}
	res, err := a.do(ctx, s, "POST", "/api/ds/query", "application/json", bytes.NewReader(body))
	if err != nil {
		return domain.PanelData{}, err
	}
	var payload queryResponse
	if err := decode(res, &payload); err != nil {
		return domain.PanelData{}, err
	}
	series := payload.series()
	if len(series) == 0 {
		return domain.PanelData{}, domain.ErrNoData
	}
	d := panel.FieldConfig.Defaults
	return domain.PanelData{Title: panel.Title, Unit: d.Unit, Decimals: d.Decimals, Series: series}, nil
}

// panel finds a panel of a dashboard by its ID, also inside rows.
func (a *API) panel(ctx context.Context, s domain.Settings, uid string, id int) (panelJSON, error) {
	var payload struct {
		Dashboard struct {
			Panels []panelJSON `json:"panels"`
		} `json:"dashboard"`
	}
	if err := a.getJSON(ctx, s, "/api/dashboards/uid/"+url.PathEscape(uid), &payload); err != nil {
		return panelJSON{}, err
	}
	var find func([]panelJSON) (panelJSON, bool)
	find = func(list []panelJSON) (panelJSON, bool) {
		for _, p := range list {
			if p.ID == id && p.Type != "row" {
				return p, true
			}
			if p, ok := find(p.Panels); ok {
				return p, true
			}
		}
		return panelJSON{}, false
	}
	if p, ok := find(payload.Dashboard.Panels); ok {
		return p, nil
	}
	return panelJSON{}, domain.ErrBadDashboard
}

// panelQueries turns the panel's visible targets into queries of
// /api/ds/query. A target without its own data source uses the panel's;
// Grafana's pseudo sources ("-- Mixed --", "-- Dashboard --") cannot be
// queried directly, so such targets are left out.
func panelQueries(p panelJSON, rng string) []map[string]json.RawMessage {
	interval, _ := json.Marshal(intervalMs(rng))
	points, _ := json.Marshal(domain.MaxPoints)
	var out []map[string]json.RawMessage
	for i, t := range p.Targets {
		if hidden(t["hide"]) {
			continue
		}
		ds := t["datasource"]
		if datasourceRef(ds) == "" {
			ds = p.Datasource
		}
		if !queryable(ds) {
			continue
		}
		q := make(map[string]json.RawMessage, len(t)+3)
		for k, v := range t {
			q[k] = v
		}
		q["datasource"] = normaliseDatasource(ds)
		if _, ok := q["refId"]; !ok {
			q["refId"], _ = json.Marshal(string(rune('A' + i%26)))
		}
		q["intervalMs"], q["maxDataPoints"] = interval, points
		out = append(out, q)
	}
	return out
}

func hidden(raw json.RawMessage) bool {
	var b bool
	return json.Unmarshal(raw, &b) == nil && b
}

// queryable reports whether a data source reference names a real source.
func queryable(raw json.RawMessage) bool {
	ref := datasourceRef(raw)
	return ref != "" && !strings.HasPrefix(ref, "-- ")
}

// datasourceRef is the uid of a reference ({"uid": …}) or the plain string
// of older dashboards; "" when there is none.
func datasourceRef(raw json.RawMessage) string {
	var obj struct {
		UID string `json:"uid"`
	}
	if json.Unmarshal(raw, &obj) == nil && obj.UID != "" {
		return obj.UID
	}
	var name string
	if json.Unmarshal(raw, &name) == nil {
		return name
	}
	return ""
}

// normaliseDatasource sends object references unchanged and wraps a plain
// string as a uid, which is how dashboards of Grafana 8+ store it.
func normaliseDatasource(raw json.RawMessage) json.RawMessage {
	var obj map[string]any
	if json.Unmarshal(raw, &obj) == nil {
		return raw
	}
	b, _ := json.Marshal(map[string]string{"uid": datasourceRef(raw)})
	return b
}

// intervalMs spreads the range over the points a widget can show.
func intervalMs(rng string) int64 {
	units := map[byte]time.Duration{'m': time.Minute, 'h': time.Hour, 'd': 24 * time.Hour, 'w': 7 * 24 * time.Hour}
	var n int64
	_, _ = fmt.Sscanf(rng[:len(rng)-1], "%d", &n)
	span := time.Duration(n) * units[rng[len(rng)-1]]
	return max(int64(span/time.Millisecond)/domain.MaxPoints, 1000)
}

// queryResponse is the answer of /api/ds/query: data frames per refId.
type queryResponse struct {
	Results map[string]struct {
		Error  string  `json:"error"`
		Frames []frame `json:"frames"`
	} `json:"results"`
}

type frame struct {
	Schema struct {
		Name   string `json:"name"`
		Fields []struct {
			Name   string            `json:"name"`
			Type   string            `json:"type"`
			Labels map[string]string `json:"labels"`
			Config struct {
				DisplayName       string `json:"displayName"`
				DisplayNameFromDS string `json:"displayNameFromDS"`
			} `json:"config"`
		} `json:"fields"`
	} `json:"schema"`
	Data struct {
		// Columns may hold strings or booleans too; only numbers are read.
		Values [][]json.RawMessage `json:"values"`
	} `json:"data"`
}

// series collects every number field of every frame, in refId order, with
// the time field as x; a frame without one (a reduced value) gets no time.
func (q queryResponse) series() []domain.Series {
	refs := make([]string, 0, len(q.Results))
	for ref := range q.Results {
		refs = append(refs, ref)
	}
	sort.Strings(refs)
	var out []domain.Series
	for _, ref := range refs {
		for _, f := range q.Results[ref].Frames {
			timeCol := -1
			for i, fld := range f.Schema.Fields {
				if fld.Type == "time" {
					timeCol = i
					break
				}
			}
			for i, fld := range f.Schema.Fields {
				if fld.Type != "number" || i >= len(f.Data.Values) {
					continue
				}
				s := domain.Series{Name: seriesName(fld.Config.DisplayNameFromDS, fld.Config.DisplayName, fld.Labels, fld.Name, f.Schema.Name, ref)}
				for j, raw := range f.Data.Values[i] {
					v, ok := number(raw)
					if !ok {
						continue
					}
					p := domain.Point{Value: v}
					if timeCol >= 0 && timeCol < len(f.Data.Values) && j < len(f.Data.Values[timeCol]) {
						if ms, ok := number(f.Data.Values[timeCol][j]); ok {
							p.Time = time.UnixMilli(int64(ms))
						}
					}
					s.Points = append(s.Points, p)
				}
				if len(s.Points) == 0 {
					continue
				}
				s.Points = domain.Downsample(s.Points, domain.MaxPoints)
				out = append(out, s)
				if len(out) == domain.MaxSeries {
					return out
				}
			}
		}
	}
	return out
}

// number reads a JSON number; null, NaN written as a string and other
// types are no value.
func number(raw json.RawMessage) (float64, bool) {
	var v float64
	return v, json.Unmarshal(raw, &v) == nil && len(raw) > 0 && raw[0] != 'n'
}

// seriesName follows Grafana's order: the name from the data source, the
// display name, the labels, then the field, frame and query names.
func seriesName(fromDS, display string, labels map[string]string, field, frameName, ref string) string {
	switch {
	case fromDS != "":
		return fromDS
	case display != "":
		return display
	case len(labels) > 0:
		keys := make([]string, 0, len(labels))
		for k := range labels {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = k + "=" + labels[k]
		}
		return strings.Join(parts, ", ")
	case field != "" && field != "Value":
		return field
	case frameName != "":
		return frameName
	}
	return ref
}

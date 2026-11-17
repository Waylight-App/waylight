package demo

import (
	"net/http"
	"strings"
	"time"
)

// Departures of the transit widget and search: the MOTIS API of Transitous.
func init() {
	handle(hostIs("api.transitous.org"), func(req *http.Request, now time.Time) (int, string, string) {
		switch req.URL.Path {
		case "/api/v1/geocode":
			return jsonOK(transitStops(req.URL.Query().Get("text")))
		case "/api/v1/stoptimes":
			return jsonOK(transitDepartures(req.URL.Query().Get("stopId"), now))
		}
		return http.StatusNotFound, "text/plain", "not found"
	})
}

// DemoStopID is the stop of the showcase's departures widget.
const DemoStopID = "demo_de:03403:16750"

type demoStop struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

var demoStops = []demoStop{
	{DemoStopID, "Oldenburg Hbf", 53.1443, 8.2227},
	{"demo_de:03403:27020", "Oldenburg(Oldb) Hbf Süd", 53.1432, 8.2216},
	{"demo_de:03403:10120", "Oldenburg Lappan", 53.1415, 8.2155},
	{"demo_de:03403:10450", "Oldenburg Julius-Mosen-Platz", 53.1389, 8.2141},
	{"demo_de:03403:11800", "Oldenburg Universität", 53.1474, 8.1803},
	{"demo_de:04011:6000", "Bremen Hbf", 53.0832, 8.8138},
}

// transitStops returns the example stops whose name contains a word of the
// query; an unknown query still finds the main station, as a real search
// would offer something close.
func transitStops(query string) string {
	q := strings.ToLower(strings.TrimSpace(query))
	var out []demoStop
	for _, s := range demoStops {
		name := strings.ToLower(s.Name)
		for _, w := range strings.Fields(q) {
			if len(w) >= 3 && strings.Contains(name, strings.TrimSuffix(w, ".")) {
				out = append(out, s)
				break
			}
		}
	}
	if len(out) == 0 {
		out = demoStops[:1]
	}
	return mustJSON(out)
}

// transitDepartures returns a plausible mix of buses and trains relative to
// now: one delayed, one cancelled.
func transitDepartures(stopID string, now time.Time) string {
	type place struct {
		Departure          string `json:"departure"`
		ScheduledDeparture string `json:"scheduledDeparture"`
		Track              string `json:"track,omitempty"`
	}
	type stopTime struct {
		Place          place  `json:"place"`
		Mode           string `json:"mode"`
		RealTime       bool   `json:"realTime"`
		Headsign       string `json:"headsign"`
		RouteShortName string `json:"routeShortName"`
		Cancelled      bool   `json:"cancelled"`
	}
	rows := []struct {
		in, delay int
		line, to  string
		mode      string
		track     string
		cancelled bool
	}{
		{2, 0, "310", "Oldenburg Lappan", "BUS", "B", false},
		{4, 0, "RE 18", "Osnabrück Hbf", "REGIONAL_RAIL", "3", false},
		{7, 3, "301", "Oldenburg Universität", "BUS", "D", false},
		{11, 0, "RS 3", "Bremen Hbf", "SUBURBAN", "1", false},
		{14, 0, "ICE 1081", "Leipzig Hbf", "HIGHSPEED_RAIL", "2", true},
		{16, 6, "IC 2036", "Norddeich Mole", "LONG_DISTANCE", "4", false},
		{21, 0, "306", "Kreyenbrück", "BUS", "A", false},
		{25, 0, "RE 1", "Hannover Hbf", "REGIONAL_RAIL", "2", false},
	}
	base := now.Truncate(time.Minute)
	out := struct {
		StopTimes []stopTime `json:"stopTimes"`
	}{}
	for _, r := range rows {
		planned := base.Add(time.Duration(r.in) * time.Minute)
		out.StopTimes = append(out.StopTimes, stopTime{
			Place: place{Departure: planned.Add(time.Duration(r.delay) * time.Minute).UTC().Format(time.RFC3339),
				ScheduledDeparture: planned.UTC().Format(time.RFC3339), Track: r.track},
			Mode: r.mode, RealTime: true, Headsign: r.to, RouteShortName: r.line, Cancelled: r.cancelled,
		})
	}
	_ = stopID // every example stop shows the same lively mix
	return mustJSON(out)
}

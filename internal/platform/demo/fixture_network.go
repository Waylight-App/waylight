package demo

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// More example services of the showcase, see fixture_homelab.go.
const (
	SonarrURL       = "http://sonarr.demo.invalid:8989"
	RadarrURL       = "http://radarr.demo.invalid:7878"
	FritzBoxURL     = "http://fritz.demo.invalid:49000"
	SpeedtestURL    = "http://speedtest.demo.invalid"
	TankerkoenigKey = "00000000-0000-4000-8000-000000000001"
)

func init() {
	handle(hostIs("sonarr.demo.invalid"), func(req *http.Request, now time.Time) (int, string, string) {
		return arr(req, now, "Sonarr")
	})
	handle(hostIs("radarr.demo.invalid"), func(req *http.Request, now time.Time) (int, string, string) {
		return arr(req, now, "Radarr")
	})
	handle(hostIs("fritz.demo.invalid"), fritzBox)
	handle(hostIs("creativecommons.tankerkoenig.de"), tankerkoenig)
	handle(hostIs("speedtest.demo.invalid"), speedtestTracker)
}

func arr(req *http.Request, now time.Time, app string) (int, string, string) {
	p := strings.TrimPrefix(strings.TrimPrefix(req.URL.Path, "/api/v3"), "/api/v1")
	day := func(d int) string { return now.Add(time.Duration(d) * 24 * time.Hour).UTC().Format(time.RFC3339) }
	switch p {
	case "/system/status":
		return jsonOK(fmt.Sprintf(`{"appName":%q,"version":"4.0.15"}`, app))
	case "/health":
		if app == "Radarr" {
			return jsonOK(`[{"type":"warning","message":"Indexer Example Tracker is unavailable due to failures for more than 6 hours"}]`)
		}
		return jsonOK(`[]`)
	case "/wanted/missing":
		if app == "Radarr" {
			return jsonOK(`{"totalRecords":2}`)
		}
		return jsonOK(`{"totalRecords":5}`)
	case "/queue":
		if app == "Radarr" {
			return jsonOK(`{"records":[{"title":"Copper.Sky.2023.1080p","status":"downloading","trackedDownloadStatus":"ok","size":4200000000,"sizeleft":1260000000,"timeleft":"00:07:40"}]}`)
		}
		return jsonOK(`{"records":[{"title":"North.Sea.Stories.S02E04.1080p","status":"downloading","trackedDownloadStatus":"ok","size":1800000000,"sizeleft":360000000,"timeleft":"00:02:10"}]}`)
	case "/calendar":
		if app == "Radarr" {
			return jsonOK(fmt.Sprintf(`[{"title":"Paper Moons","year":2026,"digitalRelease":%q,"hasFile":false},
				{"title":"The Quiet Harbour","year":2026,"inCinemas":%q,"hasFile":false}]`, day(3), day(9)))
		}
		return jsonOK(fmt.Sprintf(`[{"title":"Spring Tide","seasonNumber":2,"episodeNumber":5,"airDateUtc":%q,"series":{"title":"North Sea Stories"}},
			{"title":"The Last Ferry","seasonNumber":1,"episodeNumber":8,"airDateUtc":%q,"series":{"title":"Harbour Watch"}},
			{"title":"Gale Warning","seasonNumber":2,"episodeNumber":6,"airDateUtc":%q,"series":{"title":"North Sea Stories"}}]`, day(1), day(2), day(8)))
	}
	return http.StatusNotFound, "text/plain", "not found"
}

// fritzBox answers the UPnP and TR-064 actions of the FRITZ!Box module.
func fritzBox(req *http.Request, now time.Time) (int, string, string) {
	switch {
	case req.URL.Path == "/tr64desc.xml":
		return http.StatusOK, "text/xml", `<?xml version="1.0"?><root><device><modelName>FRITZ!Box 7590 AX</modelName></device></root>`
	case strings.HasPrefix(req.URL.Path, "/calllist.lua"):
		return http.StatusOK, "text/xml", fritzCalls(now)
	}
	action := req.Header.Get("SoapAction")
	if i := strings.LastIndex(action, "#"); i >= 0 {
		action = strings.Trim(action[i+1:], `"`)
	}
	args := map[string]string{
		"GetStatusInfo":           "<NewConnectionStatus>Connected</NewConnectionStatus><NewUptime>1033200</NewUptime>",
		"GetExternalIPAddress":    "<NewExternalIPAddress>203.0.113.42</NewExternalIPAddress>",
		"GetAddonInfos":           "<NewByteReceiveRate>4215000</NewByteReceiveRate><NewByteSendRate>612000</NewByteSendRate>",
		"GetCommonLinkProperties": "<NewLayer1DownstreamMaxBitRate>250000000</NewLayer1DownstreamMaxBitRate><NewLayer1UpstreamMaxBitRate>40000000</NewLayer1UpstreamMaxBitRate>",
		"GetInfo":                 "<NewConnectionStatus>Connected</NewConnectionStatus><NewExternalIPAddress>203.0.113.42</NewExternalIPAddress><NewUptime>1033200</NewUptime>",
		"GetCallList":             "<NewCallListURL>" + FritzBoxURL + "/calllist.lua?sid=demo</NewCallListURL>",
	}[action]
	return http.StatusOK, "text/xml", fmt.Sprintf(`<?xml version="1.0"?><s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><u:%sResponse>%s</u:%sResponse></s:Body></s:Envelope>`,
		action, args, action)
}

func fritzCalls(now time.Time) string {
	at := func(h int) string { return now.Add(-time.Duration(h) * time.Hour).Format("02.01.06 15:04") }
	calls := []struct {
		t            int
		name, number string
		hoursAgo     int
		dur          string
	}{
		{1, "Mama", "0441 123456", 1, "0:12"},
		{2, "", "0170 9876543", 3, "0:00"},
		{3, "Pizzeria Lappan", "0441 765432", 5, "0:02"},
		{1, "Praxis Dr. Meyer", "0441 555010", 26, "0:04"},
		{10, "", "0800 1112223", 30, "0:00"},
		{3, "Jonas", "0151 2233445", 49, "0:31"},
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?><root>`)
	for _, c := range calls {
		number := `<Caller>` + c.number + `</Caller><Called>0441 999888</Called>`
		if c.t == 3 {
			number = `<Caller>0441 999888</Caller><Called>` + c.number + `</Called>`
		}
		fmt.Fprintf(&b, `<Call><Type>%d</Type>%s<Name>%s</Name><Date>%s</Date><Duration>%s</Duration></Call>`,
			c.t, number, c.name, at(c.hoursAgo), c.dur)
	}
	b.WriteString(`</root>`)
	return b.String()
}

// tankerkoenig answers the fuel price API for stations around Oldenburg.
func tankerkoenig(req *http.Request, _ time.Time) (int, string, string) {
	stations := []struct {
		id, name, brand, street, nr string
		dist, e5, e10, diesel       float64
	}{
		{"0a1b2c3d-0000-4000-8000-000000000011", "Tankstelle am Stadion", "ARAL", "Maastrichter Str.", "12", 0.8, 1.799, 1.739, 1.649},
		{"0a1b2c3d-0000-4000-8000-000000000012", "Ammerländer Heerstraße", "JET", "Ammerländer Heerstr.", "201", 1.6, 1.749, 1.689, 1.609},
		{"0a1b2c3d-0000-4000-8000-000000000013", "Autohof Etzhorn", "Shell", "Butjadinger Str.", "330", 3.1, 1.829, 1.769, 1.679},
		{"0a1b2c3d-0000-4000-8000-000000000014", "Freie Tankstelle Bürgerfelde", "", "Alexanderstr.", "88", 2.2, 1.769, 1.709, 1.629},
	}
	switch req.URL.Path {
	case "/json/list.php":
		kind := req.URL.Query().Get("type")
		type st struct {
			ID          string   `json:"id"`
			Name        string   `json:"name"`
			Brand       string   `json:"brand"`
			Street      string   `json:"street"`
			HouseNumber string   `json:"houseNumber"`
			Place       string   `json:"place"`
			Dist        float64  `json:"dist"`
			Price       *float64 `json:"price"`
			IsOpen      bool     `json:"isOpen"`
		}
		var out []st
		for _, s := range stations {
			price := s.e5
			switch kind {
			case "e10":
				price = s.e10
			case "diesel":
				price = s.diesel
			}
			out = append(out, st{s.id, s.name, s.brand, s.street, s.nr, "Oldenburg", s.dist, &price, true})
		}
		return jsonOK(mustJSON(map[string]any{"ok": true, "status": "ok", "stations": out}))
	case "/json/prices.php":
		prices := map[string]any{}
		for _, s := range stations {
			prices[s.id] = map[string]any{"status": "open", "e5": s.e5, "e10": s.e10, "diesel": s.diesel}
		}
		return jsonOK(mustJSON(map[string]any{"ok": true, "prices": prices}))
	}
	return http.StatusNotFound, "text/plain", "not found"
}

// speedtestTracker answers the results API with a measurement every hour of
// the last two days.
func speedtestTracker(_ *http.Request, now time.Time) (int, string, string) {
	type result struct {
		Ping         float64 `json:"ping"`
		Download     float64 `json:"download"`
		Upload       float64 `json:"upload"`
		DownloadBits float64 `json:"download_bits"`
		UploadBits   float64 `json:"upload_bits"`
		Status       string  `json:"status"`
		CreatedAt    string  `json:"created_at"`
	}
	var data []result
	base := now.Truncate(time.Hour)
	for i := range 48 {
		// Evenings are a little slower, like on a real cable line.
		at := base.Add(-time.Duration(i) * time.Hour)
		evening := at.Hour() >= 19 && at.Hour() <= 22
		down := 236.0 - float64(i%5)*3.1
		if evening {
			down -= 41
		}
		up := 38.5 - float64(i%3)*0.9
		data = append(data, result{Ping: 11 + float64(i%4), Download: down * 125_000, Upload: up * 125_000,
			DownloadBits: down * 1e6, UploadBits: up * 1e6, Status: "completed", CreatedAt: at.UTC().Format("2006-01-02 15:04:05")})
	}
	return jsonOK(mustJSON(map[string]any{"data": data}))
}

package demo

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Example homelab services of the showcase. Their host names end in
// .demo.invalid, a reserved name that never resolves, so only these fixtures
// can answer them.
const (
	AdGuardURL       = "http://adguard.demo.invalid"
	JellyfinURL      = "http://jellyfin.demo.invalid:8096"
	ProxmoxURL       = "https://pve.demo.invalid:8006"
	ImmichURL        = "http://photos.demo.invalid:2283"
	HomeAssistantURL = "http://home.demo.invalid:8123"
	NextcloudURL     = "https://cloud.demo.invalid"
)

func init() {
	handle(hostIs("adguard.demo.invalid"), adguard)
	handle(hostIs("jellyfin.demo.invalid"), jellyfin)
	handle(hostIs("pve.demo.invalid"), proxmox)
	handle(hostIs("photos.demo.invalid"), immich)
	handle(hostIs("home.demo.invalid"), homeAssistant)
	handle(hostIs("cloud.demo.invalid"), nextcloud)
}

func adguard(req *http.Request, _ time.Time) (int, string, string) {
	switch req.URL.Path {
	case "/control/stats":
		return jsonOK(`{"num_dns_queries":48213,"num_blocked_filtering":9127,"avg_processing_time":0.0042,
			"top_blocked_domains":[{"telemetry.example-tv.com":1844},{"ads.example-cdn.net":1201},{"tracker.example-shop.de":766},{"metrics.example-app.io":512}]}`)
	case "/control/status":
		return jsonOK(`{"protection_enabled":true,"protection_disabled_duration":0}`)
	case "/control/protection":
		return jsonOK(`{}`)
	}
	return http.StatusNotFound, "text/plain", "not found"
}

func jellyfin(req *http.Request, now time.Time) (int, string, string) {
	p := req.URL.Path
	switch {
	case p == "/System/Info":
		return jsonOK(`{"ServerName":"Living room media server"}`)
	case p == "/Items":
		type item struct {
			ID             string `json:"Id"`
			Name           string `json:"Name"`
			SeriesName     string `json:"SeriesName,omitempty"`
			AlbumArtist    string `json:"AlbumArtist,omitempty"`
			Type           string `json:"Type"`
			ProductionYear int    `json:"ProductionYear,omitempty"`
			DateCreated    string `json:"DateCreated"`
			RunTimeTicks   int64  `json:"RunTimeTicks"`
		}
		ago := func(h int) string { return now.Add(-time.Duration(h) * time.Hour).UTC().Format(time.RFC3339) }
		items := []item{
			{"a1", "The Lighthouse Keeper", "", "", "Movie", 2025, ago(3), 64_800_000_000},
			{"a2", "Low Tide", "North Sea Stories", "", "Episode", 2026, ago(9), 27_000_000_000},
			{"a3", "Harbour Lights", "", "Nordlicht Ensemble", "MusicAlbum", 2026, ago(20), 26_400_000_000},
			{"a4", "Paper Moons", "", "", "Movie", 2024, ago(30), 70_200_000_000},
			{"a5", "The Wadden Line", "North Sea Stories", "", "Episode", 2026, ago(50), 27_600_000_000},
			{"a6", "Copper Sky", "", "", "Movie", 2023, ago(75), 58_200_000_000},
		}
		return jsonOK(mustJSON(map[string]any{"Items": items}))
	case p == "/Sessions":
		return jsonOK(`[{"UserName":"Anna","DeviceName":"Living room TV","NowPlayingItem":{"Id":"a1","Name":"The Lighthouse Keeper","Type":"Movie","ProductionYear":2025,"RunTimeTicks":64800000000},"PlayState":{"PositionTicks":25200000000,"IsPaused":false}}]`)
	case strings.HasSuffix(p, "/Images/Primary"):
		return http.StatusOK, "image/png", picture(p, 120, 180)
	}
	return http.StatusNotFound, "text/plain", "not found"
}

func proxmox(req *http.Request, _ time.Time) (int, string, string) {
	if req.URL.Path != "/api2/json/cluster/resources" {
		// Start and stop of a guest: Proxmox answers with a task ID.
		return jsonOK(`{"data":"UPID:pve:000F4240:demo:start:qemu::root@pam:"}`)
	}
	const gib = int64(1) << 30
	type res struct {
		ID      string  `json:"id"`
		Type    string  `json:"type"`
		Node    string  `json:"node"`
		Status  string  `json:"status"`
		Name    string  `json:"name,omitempty"`
		VMID    int     `json:"vmid,omitempty"`
		Storage string  `json:"storage,omitempty"`
		CPU     float64 `json:"cpu"`
		MaxCPU  float64 `json:"maxcpu"`
		Mem     int64   `json:"mem"`
		MaxMem  int64   `json:"maxmem"`
		Disk    int64   `json:"disk"`
		MaxDisk int64   `json:"maxdisk"`
		Uptime  int64   `json:"uptime"`
	}
	guest := func(kind string, id int, name, status string, cpu float64, mem, maxMem int64) res {
		up := int64(0)
		if status == "running" {
			up = 864_000 + int64(id)*3_600
		}
		return res{ID: fmt.Sprintf("%s/%d", kind, id), Type: kind, Node: "pve", Status: status, Name: name, VMID: id,
			CPU: cpu, MaxCPU: 2, Mem: mem, MaxMem: maxMem, Disk: 6 * gib, MaxDisk: 32 * gib, Uptime: up}
	}
	data := []res{
		{ID: "node/pve", Type: "node", Node: "pve", Status: "online", CPU: 0.18, MaxCPU: 8, Mem: 21 * gib, MaxMem: 64 * gib,
			Disk: 38 * gib, MaxDisk: 94 * gib, Uptime: 1_728_000},
		guest("qemu", 100, "home-assistant", "running", 0.06, 3*gib, 4*gib),
		guest("qemu", 105, "windows-11", "stopped", 0, 0, 8*gib),
		guest("lxc", 101, "pihole", "running", 0.01, gib/4, gib/2),
		guest("lxc", 102, "jellyfin", "running", 0.22, 3*gib, 6*gib),
		guest("lxc", 103, "paperless", "running", 0.04, 2*gib, 4*gib),
		guest("lxc", 104, "nextcloud", "running", 0.09, 4*gib, 8*gib),
		{ID: "storage/pve/local-lvm", Type: "storage", Node: "pve", Status: "available", Storage: "local-lvm",
			Disk: 412 * gib, MaxDisk: 900 * gib},
	}
	return jsonOK(mustJSON(map[string]any{"data": data}))
}

// immichAssets are UUIDs, the format Immich uses for asset IDs.
var immichAssets = []string{
	"6f1c2a3e-4b5d-4e6f-8a9b-0c1d2e3f4a5b", "7a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
	"8b3c4d5e-6f7a-4b8c-9d0e-1f2a3b4c5d6e", "9c4d5e6f-7a8b-4c9d-8e1f-2a3b4c5d6e7f",
}

func immich(req *http.Request, now time.Time) (int, string, string) {
	p := req.URL.Path
	switch {
	case p == "/api/server/version":
		return jsonOK(`{"major":2,"minor":1,"patch":0}`)
	case p == "/api/assets/statistics":
		return jsonOK(`{"images":18432,"videos":1206,"total":19638}`)
	case p == "/api/server/storage":
		return jsonOK(`{"diskSizeRaw":2000398934016,"diskUseRaw":812345678848}`)
	case p == "/api/memories":
		type asset struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}
		mem := func(yearsAgo int, ids ...string) map[string]any {
			as := make([]asset, 0, len(ids))
			for _, id := range ids {
				as = append(as, asset{id, "IMAGE"})
			}
			return map[string]any{"type": "on_this_day", "data": map[string]int{"year": now.Year() - yearsAgo}, "assets": as}
		}
		return jsonOK(mustJSON([]any{mem(1, immichAssets[0], immichAssets[1]), mem(4, immichAssets[2], immichAssets[3])}))
	case strings.HasPrefix(p, "/api/assets/"):
		return http.StatusOK, "image/png", picture(p, 320, 240)
	}
	return http.StatusNotFound, "text/plain", "not found"
}

func homeAssistant(req *http.Request, now time.Time) (int, string, string) {
	if req.URL.Path != "/api/states" {
		// Toggling a switch: Home Assistant answers with the changed states.
		return jsonOK(`[]`)
	}
	type attrs struct {
		FriendlyName string `json:"friendly_name"`
		Unit         string `json:"unit_of_measurement,omitempty"`
	}
	type entity struct {
		EntityID    string `json:"entity_id"`
		State       string `json:"state"`
		Attributes  attrs  `json:"attributes"`
		LastChanged string `json:"last_changed"`
	}
	ago := func(m int) string { return now.Add(-time.Duration(m) * time.Minute).UTC().Format(time.RFC3339) }
	states := []entity{
		{"sensor.living_room_temperature", "21.4", attrs{"Living room", "°C"}, ago(4)},
		{"sensor.balcony_solar_power", "412", attrs{"Balcony solar", "W"}, ago(1)},
		{"sensor.washing_machine_power", "1.2", attrs{"Washing machine", "W"}, ago(35)},
		{"switch.coffee_machine", "off", attrs{"Coffee machine", ""}, ago(80)},
		{"light.desk_lamp", "on", attrs{"Desk lamp", ""}, ago(12)},
		{"binary_sensor.front_door", "off", attrs{"Front door", ""}, ago(140)},
	}
	return jsonOK(mustJSON(states))
}

func nextcloud(req *http.Request, _ time.Time) (int, string, string) {
	p := req.URL.Path
	switch {
	case p == "/status.php":
		return jsonOK(`{"installed":true,"maintenance":false,"versionstring":"31.0.4"}`)
	case strings.HasPrefix(p, "/ocs/v1.php/cloud/users/"):
		return jsonOK(`{"ocs":{"meta":{"statuscode":100},"data":{"quota":{"used":96636764160,"free":118111600640,"quota":214748364800}}}}`)
	case strings.HasPrefix(p, "/ocs/v2.php/apps/serverinfo/"):
		return jsonOK(`{"ocs":{"meta":{"statuscode":200},"data":{"nextcloud":{"system":{"freespace":1288490188800,"apps":{"num_updates_available":2}},
			"storage":{"num_users":4,"num_files":214806}},"server":{"database":{"size":"912261120"}},"activeUsers":{"last5minutes":1,"last24hours":3}}}}`)
	}
	return http.StatusNotFound, "text/plain", "not found"
}

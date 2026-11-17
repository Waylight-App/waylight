package usecase

// ShowcaseLayout is the layout of the public demo (DEMO_MODE): it shows as
// many widgets as possible, each with sample data, so visitors see what the
// product can do. weatherConfig is the JSON configuration of the weather
// widget.
func ShowcaseLayout(weatherConfig string) []SampleTab {
	grid := func(kinds ...string) []SampleWidget {
		out := make([]SampleWidget, 0, len(kinds))
		for i, k := range kinds {
			cfg := showcaseConfig(k)
			if k == "weather" {
				cfg = weatherConfig
			}
			out = append(out, SampleWidget{Kind: k, X: (i % 4) * 3, Y: (i / 4) * 5, W: 3, H: 5, Config: cfg})
		}
		return out
	}
	return []SampleTab{
		{Name: "Home", Widgets: grid("clock", "weather", "todos", "notes", "calendar", "bookmarks", "rss", "countdown",
			"habits", "focus", "birthdays", "holidays", "briefing", "timeline", "study", "readlater", "shopping", "drop")},
		{Name: "Out and about", Widgets: grid("transit", "waste", "parcels", "fuel", "radar", "space", "sports", "markets",
			"electricity", "gamedeals", "photo-frame", "immich", "media")},
		{Name: "Server", Widgets: grid("services", "status-overview", "system", "docker", "adguard", "diskhealth", "speedtest",
			"backup-status", "proxmox", "wireguard", "releases", "nextcloud", "fritzbox", "fritzbox-calls", "arr", "apiwidget",
			"mail", "smarthome", "phonelist")},
	}
}

// showcaseConfig is the widget configuration of a kind in the showcase, from
// the local examples or the example integrations; "" uses the defaults.
func showcaseConfig(kind string) string {
	if c, ok := localShowcaseConfigs[kind]; ok {
		return c
	}
	return integrationShowcaseConfigs[kind]
}

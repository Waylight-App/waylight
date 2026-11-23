package usecase

// ShowcaseLayout is the layout of the public demo (DEMO_MODE): it shows as
// many widgets as possible, each with sample data, so visitors see what the
// product can do. weatherConfig is the JSON configuration of the weather
// card.
func ShowcaseLayout(weatherConfig string) []SampleTab {
	// Each tab opens with wide widgets (#769); the rest are equal three
	// column tiles below, so the demo still shows as many widgets as possible.
	tab := func(name string, lead []SampleWidget, kinds ...string) SampleTab {
		widgets := lead
		top := 0
		if len(lead) > 0 {
			top = 5
		}
		for i, k := range kinds {
			cfg := showcaseConfig(k)
			if k == "weather" {
				cfg = weatherConfig
			}
			widgets = append(widgets, SampleWidget{Kind: k, X: (i % 4) * 3, Y: top + (i/4)*5, W: 3, H: 5, Config: cfg})
		}
		return SampleTab{Name: name, Widgets: widgets}
	}
	wide := func(kind string, x, w int, config string) SampleWidget {
		return SampleWidget{Kind: kind, X: x, Y: 0, W: w, H: 5, Config: config}
	}
	return []SampleTab{
		tab("Home", []SampleWidget{
			wide("tonight", 0, 6, weatherConfig), wide("todos", 6, 3, showcaseConfig("todos")), wide("notes", 9, 3, showcaseConfig("notes")),
		}, "clock", "weather", "calendar", "bookmarks", "rss", "countdown",
			"habits", "focus", "birthdays", "holidays", "briefing", "timeline", "study", "readlater", "shopping", "drop"),
		tab("Out and about", nil, "transit", "waste", "parcels", "fuel", "radar", "space", "sports", "markets",
			"electricity", "gamedeals", "photo-frame", "immich", "media"),
		tab("Server", []SampleWidget{
			wide("system", 0, 8, showcaseConfig("system")), wide("services", 8, 4, showcaseConfig("services")),
		}, "status-overview", "docker", "adguard", "diskhealth", "speedtest",
			"backup-status", "proxmox", "wireguard", "releases", "nextcloud", "fritzbox", "fritzbox-calls", "arr", "apiwidget",
			"smarthome", "phonelist", "grafana-panel", "grafana-alerts"),
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

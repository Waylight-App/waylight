package usecase

// ShowcaseLayout is the layout of the public demo (DEMO_MODE): it shows as
// many widgets as possible, each with sample data, so visitors see what the
// product can do. weatherConfig is the JSON configuration of the weather
// card.
func ShowcaseLayout(weatherConfig string) []SampleTab {
	// Each tab opens with the composition of the design mockups (#796): a lead
	// row, then a row of three four-column widgets; the rest are equal
	// three-column tiles below, so the demo still shows as many widgets as
	// possible. Every row adds up to the 12 columns.
	type lead struct {
		kind string
		w    int
	}
	tab := func(name string, leads []lead, mid []string, kinds ...string) SampleTab {
		var widgets []SampleWidget
		x, y := 0, 0
		for _, l := range leads {
			cfg := showcaseConfig(l.kind)
			if l.kind == "tonight" {
				cfg = weatherConfig
			}
			widgets = append(widgets, SampleWidget{Kind: l.kind, X: x, Y: y, W: l.w, H: sampleRowHeight, Config: cfg})
			x += l.w
		}
		if len(leads) > 0 {
			y += sampleRowHeight
		}
		for i, k := range mid {
			widgets = append(widgets, SampleWidget{Kind: k, X: (i % 3) * 4, Y: y + (i/3)*sampleRowHeight, W: 4, H: sampleRowHeight, Config: showcaseConfig(k)})
		}
		y += ((len(mid) + 2) / 3) * sampleRowHeight
		for i, k := range kinds {
			cfg := showcaseConfig(k)
			if k == "weather" {
				cfg = weatherConfig
			}
			widgets = append(widgets, SampleWidget{Kind: k, X: (i % 4) * 3, Y: y + (i/4)*sampleRowHeight, W: 3, H: sampleRowHeight, Config: cfg})
		}
		return SampleTab{Name: name, Widgets: widgets}
	}
	return []SampleTab{
		tab("Home", []lead{{"tonight", 5}, {"todos", 4}, {"notes", 3}},
			[]string{"calendar", "bookmarks", "rss", "habits", "birthdays", "focus"},
			"clock", "weather", "countdown", "holidays", "briefing", "timeline", "study", "readlater", "shopping", "drop"),
		tab("Out and about", nil, nil, "transit", "waste", "parcels", "fuel", "radar", "space", "sports", "markets",
			"electricity", "gamedeals", "photo-frame", "immich", "media"),
		tab("Server", []lead{{"speedtest", 8}, {"backup-status", 4}},
			[]string{"system", "docker", "diskhealth"},
			"services", "status-overview", "adguard", "proxmox", "wireguard", "releases", "nextcloud", "fritzbox", "fritzbox-calls",
			"arr", "apiwidget", "smarthome", "phonelist", "grafana-panel", "grafana-alerts"),
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

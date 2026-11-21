package grafana

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "grafana-panel",
			Shows: kernel.NewText("One panel of a Grafana dashboard, as a value or a small chart."),
			SetUp: kernel.NewText("Choose Set up in the widget to connect Grafana, then pick the dashboard and the panel."),
			Docs:  "modules/grafana/",
		},
		{
			Kind:  "grafana-alerts",
			Shows: kernel.NewText("The alerts that are firing in Grafana right now."),
			SetUp: kernel.NewText("It uses the Grafana connection; choose Set up in the widget if there is none yet."),
			Docs:  "modules/grafana/",
		},
	}
}

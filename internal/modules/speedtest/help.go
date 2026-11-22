package speedtest

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "speedtest",
			Shows: kernel.NewText("Your internet speed over time, from Speedtest Tracker."),
			SetUp: kernel.NewText("Choose Connect in the widget and enter the address of your Speedtest Tracker and its access key."),
			Docs:  "features/#network-and-integrations",
		},
	}
}

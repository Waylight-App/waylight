package transit

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "transit",
			Shows: kernel.NewText("The next departures from a bus or train stop, with delays."),
			SetUp: kernel.NewText("Open the widget settings and search for your stop."),
			Docs:  "features/#information",
		},
	}
}

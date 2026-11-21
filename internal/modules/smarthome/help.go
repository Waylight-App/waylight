package smarthome

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "smarthome",
			Shows: kernel.NewText("Devices from Home Assistant as tiles, with their state and, if allowed, a switch."),
			SetUp: kernel.NewText("Choose Set up in the widget and sign in with Home Assistant, then pick the devices."),
			Docs:  "host-integrations/#home-assistant",
		},
	}
}

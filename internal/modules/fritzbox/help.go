package fritzbox

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "fritzbox",
			Shows: kernel.NewText("Your internet connection as the FRITZ!Box sees it: online status, speed and how long it has been up."),
			SetUp: kernel.NewText("Choose Set up in the widget and enter the address of your FRITZ!Box and a user name and password of the box."),
			Docs:  "features/#network-and-integrations",
		},
		{
			Kind:  "fritzbox-calls",
			Shows: kernel.NewText("Recent calls of your FRITZ!Box: missed, incoming and outgoing."),
			SetUp: kernel.NewText("It uses the FRITZ!Box you set up for the internet widget. Open the widget settings to choose which calls are shown."),
			Docs:  "features/#network-and-integrations",
		},
	}
}

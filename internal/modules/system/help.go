package system

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "system",
			Shows: kernel.NewText("How busy the server is: processor, memory, disks, temperature and how long it has been running."),
			SetUp: kernel.NewText("Nothing to set up; it shows the machine Waylight runs on."),
			Docs:  "features/#home-server",
		},
	}
}

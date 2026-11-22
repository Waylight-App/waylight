package proxmox

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "proxmox",
			Shows: kernel.NewText("Your Proxmox server: load and which virtual machines and containers are running."),
			SetUp: kernel.NewText("Choose Connect in the widget and enter the Proxmox address and a read-only access key."),
			Docs:  "host-integrations/#proxmox-ve",
		},
	}
}

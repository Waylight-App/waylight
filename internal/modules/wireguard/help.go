package wireguard

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "wireguard",
			Shows: kernel.NewText("The devices of your WireGuard VPN with their last connection and traffic."),
			SetUp: kernel.NewText("An admin lets the server write a small status file for Waylight; the guide explains the few steps."),
			Docs:  "host-integrations/#wireguard",
		},
	}
}

package services

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "services",
			Shows: kernel.NewText("Tiles for your own services with their online status, response time and certificate warnings."),
			SetUp: kernel.NewText("Add services in Settings under Services, or let Waylight find them on your network. Containers with Waylight labels appear by themselves."),
			Docs:  "host-integrations/#docker-autodiscovery",
		},
		{
			Kind:  "status-overview",
			Shows: kernel.NewText("A status page of your services: uptime over the last days and current or past outages."),
			SetUp: kernel.NewText("It uses the services you added in Settings under Services; there is nothing else to set up."),
			Docs:  "features/#home-server",
		},
	}
}

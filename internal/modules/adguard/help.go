package adguard

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "adguard",
			Shows: kernel.NewText("How many requests AdGuard Home or Pi-hole blocked today, and whether blocking is on."),
			SetUp: kernel.NewText("Choose Connect in the widget and enter the address and password of your AdGuard Home or Pi-hole."),
			Docs:  "host-integrations/#dns-filter-pause",
		},
	}
}

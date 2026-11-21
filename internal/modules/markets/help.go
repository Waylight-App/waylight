package markets

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "markets",
			Shows: kernel.NewText("Currencies, shares, indices and crypto you watch, with today's change."),
			SetUp: kernel.NewText("Open the widget settings and add what you want to follow. No account or key is needed."),
			Docs:  "modules/markets/",
		},
	}
}

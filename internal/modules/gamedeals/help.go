package gamedeals

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "gamedeals",
			Shows: kernel.NewText("Discounts on games on your Steam wishlist and free games on the Epic Games Store."),
			SetUp: kernel.NewText("The free games need no setup. For wishlist discounts, enter your Steam profile in Settings under Game deals."),
			Docs:  "modules/gamedeals/",
		},
	}
}

package fuel

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "fuel",
			Shows: kernel.NewText("The cheapest open petrol stations near a place in Germany, or your favourites."),
			SetUp: kernel.NewText("Choose Connect in the widget and enter your free key from Tankerkönig, then pick a place."),
			Docs:  "modules/fuel/",
		},
	}
}

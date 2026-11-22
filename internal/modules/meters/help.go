package meters

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "meters",
			Shows: kernel.NewText("Your electricity, gas, water and heat meters with the readings, the consumption per month and the consumption per year."),
			SetUp: kernel.NewText("Add a meter with its name and unit, then enter a reading with the day and the counter value. Enter one regularly, for example on the first of the month."),
			Docs:  "modules/meters/",
		},
	}
}

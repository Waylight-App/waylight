package meals

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "meals",
			Shows: kernel.NewText("A week of dishes with recipe links and ingredients."),
			SetUp: kernel.NewText("Plan a dish with the plus next to a day and list its ingredients. With the shopping list on, one click adds them to it."),
			Docs:  "modules/meals/",
		},
	}
}

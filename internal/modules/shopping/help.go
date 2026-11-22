package shopping

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "shopping",
			Shows: kernel.NewText("A shopping list you can share with your household."),
			SetUp: kernel.NewText("Add items right in the widget. Choose Connect a list to use a task list from your calendar server instead."),
			Docs:  "modules/shopping/",
		},
	}
}

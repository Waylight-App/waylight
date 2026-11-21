package chores

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "chores",
			Shows: kernel.NewText("Recurring household chores that the people on the list take turns with."),
			SetUp: kernel.NewText("Add a chore with New chore, set how often it repeats and list the people who take turns. Tick it off when it is done."),
			Docs:  "modules/chores/",
		},
	}
}

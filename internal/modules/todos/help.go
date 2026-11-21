package todos

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "todos",
			Shows: kernel.NewText("Your to-dos with due dates, or a task list from Nextcloud, Todoist or Microsoft To Do."),
			SetUp: kernel.NewText("Type \"t\" and a task into the search to add one. Open the widget settings to show a connected task list instead."),
			Docs:  "modules/todos/",
		},
	}
}

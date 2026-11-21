package absences

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "absences",
			Shows: kernel.NewText("Who is away: on holiday, sick, working remotely or away for another reason."),
			SetUp: kernel.NewText("Enter your own absences in the widget with the first and last day; admins can enter them for anyone."),
			Docs:  "absences/",
		},
	}
}

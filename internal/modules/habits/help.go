package habits

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "habits",
			Shows: kernel.NewText("Your daily habits with streaks: tick off what you did today."),
			SetUp: kernel.NewText("Add habits in the widget and choose the days they are due; tap a habit to tick it off."),
			Docs:  "modules/habits/",
		},
	}
}

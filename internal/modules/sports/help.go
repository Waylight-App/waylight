package sports

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "sports",
			Shows: kernel.NewText("The last result, next match and table position of your team, or the top of the table."),
			SetUp: kernel.NewText("Open the widget settings and choose a league and a team."),
			Docs:  "modules/sports/",
		},
	}
}

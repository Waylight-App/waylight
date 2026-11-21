package phonelist

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "phonelist",
			Shows: kernel.NewText("A phone list of names, roles and numbers that you can search."),
			SetUp: kernel.NewText("An admin adds the entries in Settings under Phone list. Open the widget settings to group the list by team."),
			Docs:  "modules/phonelist/",
		},
	}
}

package holidays

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "holidays",
			Shows: kernel.NewText("The next public holidays and school holidays."),
			SetUp: kernel.NewText("Open the widget settings and choose your German state. Add a calendar link for school holidays if you like."),
			Docs:  "modules/holidays/",
		},
	}
}

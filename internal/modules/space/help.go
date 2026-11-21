package space

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "space",
			Shows: kernel.NewText("The next rocket launches and NASA's astronomy picture of the day."),
			SetUp: kernel.NewText("Turn on launches or the picture in Settings under Space; nothing is fetched until then."),
			Docs:  "modules/space/",
		},
	}
}

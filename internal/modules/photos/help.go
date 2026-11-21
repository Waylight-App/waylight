package photos

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "photo-frame",
			Shows: kernel.NewText("Your own pictures, one after another."),
			SetUp: kernel.NewText("Open the widget settings to upload pictures, choose the album and how long each picture stays."),
			Docs:  "modules/photos/",
		},
	}
}

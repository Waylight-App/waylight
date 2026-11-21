package immich

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "immich",
			Shows: kernel.NewText("Your Immich photo library and the photos taken on this day in past years."),
			SetUp: kernel.NewText("Choose Set up in the widget and enter the address of Immich and a key you create in its settings."),
			Docs:  "modules/immich/",
		},
	}
}

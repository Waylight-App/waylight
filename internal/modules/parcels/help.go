package parcels

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "parcels",
			Shows: kernel.NewText("Your parcels and where they are right now."),
			SetUp: kernel.NewText("Add a tracking number with the plus button in the widget. For live DHL status, an admin adds a free DHL key in Settings under Parcels."),
			Docs:  "parcels-mail/",
		},
	}
}

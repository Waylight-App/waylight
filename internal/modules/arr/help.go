package arr

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "arr",
			Shows: kernel.NewText("Upcoming episodes and films, downloads in progress and warnings from Sonarr, Radarr or Prowlarr."),
			SetUp: kernel.NewText("Choose Set up in the widget, pick the app and enter its address and key."),
			Docs:  "modules/arr/",
		},
	}
}

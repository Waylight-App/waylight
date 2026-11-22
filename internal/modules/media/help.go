package media

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "media",
			Shows: kernel.NewText("What is playing on Jellyfin or Plex right now and what was added recently."),
			SetUp: kernel.NewText("Choose Connect in the widget, pick Jellyfin or Plex and enter its address and key."),
			Docs:  "modules/media/",
		},
	}
}

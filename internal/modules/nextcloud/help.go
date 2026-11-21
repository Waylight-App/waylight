package nextcloud

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "nextcloud",
			Shows: kernel.NewText("Your Nextcloud: version, your storage, and for admins free space, users and app updates."),
			SetUp: kernel.NewText("Choose Set up in the widget and sign in with Nextcloud, or enter its address and an app password."),
			Docs:  "features/#network-and-integrations",
		},
	}
}

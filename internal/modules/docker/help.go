package docker

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "docker",
			Shows: kernel.NewText("The containers on your server and whether they are running and healthy."),
			SetUp: kernel.NewText("An admin connects Waylight to Docker once, as the installation guide describes. Open the widget settings to choose which containers are shown."),
			Docs:  "host-integrations/#docker-actions",
		},
	}
}

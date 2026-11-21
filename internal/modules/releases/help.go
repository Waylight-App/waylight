package releases

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "releases",
			Shows: kernel.NewText("New albums and singles of the artists you follow."),
			SetUp: kernel.NewText("Follow artists in Settings under New releases."),
			Docs:  "modules/releases/",
		},
	}
}

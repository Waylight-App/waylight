package gifts

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "gifts",
			Shows: kernel.NewText("Your private gift ideas per person, with the next birthday first."),
			SetUp: kernel.NewText("Add an idea with a person and a title. Nobody else sees your ideas, not even an administrator."),
			Docs:  "modules/gifts/",
		},
	}
}

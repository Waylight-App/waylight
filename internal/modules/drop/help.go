package drop

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "drop",
			Shows: kernel.NewText("Text, links and files you send between your devices; they disappear after a while."),
			SetUp: kernel.NewText("Drop something into the widget on one device and open it on another. There is nothing to set up."),
			Docs:  "modules/drop/",
		},
	}
}

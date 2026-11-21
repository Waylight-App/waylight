package readlater

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "readlater",
			Shows: kernel.NewText("Pages you saved to read later, kept as a clean copy."),
			SetUp: kernel.NewText("Paste a link into the widget or save it with the browser extension."),
			Docs:  "modules/readlater/",
		},
	}
}

package notes

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "notes",
			Shows: kernel.NewText("One note or your pinned notes, written in Markdown."),
			SetUp: kernel.NewText("Type \"n\" and some text into the search to add a note, then pin it. Open the widget settings to choose which note it shows."),
			Docs:  "features/#everyday",
		},
	}
}

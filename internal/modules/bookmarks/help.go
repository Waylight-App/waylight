package bookmarks

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "bookmarks",
			Shows: kernel.NewText("Your links, sorted into folders and tags."),
			SetUp: kernel.NewText("Add links with the plus button or type \"b\" and a link into the search. Open the widget settings to show only one folder."),
			Docs:  "features/#everyday",
		},
	}
}

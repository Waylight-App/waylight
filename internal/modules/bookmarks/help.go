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
		{
			Kind:  "quicklinks",
			Shows: kernel.NewText("Your favourite bookmarks as tiles; the keys 1 to 9 open the first nine."),
			SetUp: kernel.NewText("Open the widget settings to choose up to twelve bookmarks and their order. Without a choice it shows your first bookmarks."),
			Docs:  "features/#everyday",
		},
	}
}

package rss

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "rss",
			Shows: kernel.NewText("The newest articles of the news feeds you follow."),
			SetUp: kernel.NewText("Add feeds in Settings under Feeds, or pick news sources from the built-in catalogue there."),
			Docs:  "features/#information",
		},
	}
}

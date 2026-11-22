package board

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "board",
			Shows: kernel.NewText("Short notes for the household, pinned or with a last day."),
			SetUp: kernel.NewText("Write a note with New note. Pin the ones that should stay on top and give short-lived ones a last day."),
			Docs:  "modules/board/",
		},
	}
}

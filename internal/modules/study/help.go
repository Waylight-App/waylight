package study

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "study",
			Shows: kernel.NewText("Your progress through your studies: credits, semesters and average grade."),
			SetUp: kernel.NewText("Add your courses in Settings under Study progress."),
			Docs:  "modules/study/",
		},
	}
}

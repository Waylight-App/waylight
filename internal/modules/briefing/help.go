package briefing

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "briefing",
			Shows: kernel.NewText("A short morning summary: weather warnings, problems overnight, due to-dos, appointments and waste collection."),
			SetUp: kernel.NewText("It collects what the other widgets know. In Settings under Daily briefing you can have it sent to you each morning."),
			Docs:  "features/#information",
		},
	}
}

package presence

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "presence",
			Shows: kernel.NewText("Who is in the office, working from home or away today."),
			SetUp: kernel.NewText("Set your own status in the widget: Office, Home or Away. It lapses at midnight."),
			Docs:  "presence/",
		},
	}
}

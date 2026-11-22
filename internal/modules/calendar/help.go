package calendar

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "calendar",
			Shows: kernel.NewText("Your appointments as an agenda, a month or both side by side."),
			SetUp: kernel.NewText("Choose Connect in the widget to connect Nextcloud or another calendar server, or add a Google or Microsoft account or a calendar link in Settings under Calendar."),
			Docs:  "modules/calendar/",
		},
	}
}

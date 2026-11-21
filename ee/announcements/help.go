package announcements

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "announcements",
			Shows: kernel.NewText("Short notices for everyone, such as \"The office is closed on Friday\"."),
			SetUp: kernel.NewText("Admins and the chosen writer group post announcements in the widget; everyone else reads them."),
			Docs:  "announcements/",
		},
	}
}

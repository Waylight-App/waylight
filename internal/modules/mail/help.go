package mail

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "mail",
			Shows: kernel.NewText("How many unread e-mails you have and the newest ones. Waylight only reads, it never sends or deletes."),
			SetUp: kernel.NewText("Add a mailbox in Settings under Mail with the server name and your login."),
			Docs:  "modules/mail/",
		},
	}
}

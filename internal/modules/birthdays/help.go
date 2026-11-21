package birthdays

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "birthdays",
			Shows: kernel.NewText("The next birthdays and how old everyone turns."),
			SetUp: kernel.NewText("Add birthdays in Settings under Birthdays, import a contacts file or connect an address book."),
			Docs:  "modules/birthdays/",
		},
	}
}

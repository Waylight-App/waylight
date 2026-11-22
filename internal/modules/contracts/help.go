package contracts

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "contracts",
			Shows: kernel.NewText("Your contracts by next cancellation deadline, with the monthly total cost."),
			SetUp: kernel.NewText("Add a contract with its cost, the end of the term and the notice period. You get a reminder before the cancellation deadline."),
			Docs:  "modules/contracts/",
		},
	}
}

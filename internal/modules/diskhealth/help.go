package diskhealth

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "diskhealth",
			Shows: kernel.NewText("The health of your drives: overall verdict, temperature and wear."),
			SetUp: kernel.NewText("An admin lets the server write a small drive report for Waylight; the guide explains how."),
			Docs:  "host-integrations/#disk-health",
		},
	}
}

package waste

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "waste",
			Shows: kernel.NewText("The next waste collection dates; the bin due tomorrow is highlighted."),
			SetUp: kernel.NewText("In Settings under Waste, add the calendar link of your town or upload its calendar file."),
			Docs:  "features/#information",
		},
	}
}

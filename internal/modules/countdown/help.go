package countdown

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "countdown",
			Shows: kernel.NewText("How many days are left until a date, such as a holiday or a deadline."),
			SetUp: kernel.NewText("Open the widget settings and add a date with a name."),
			Docs:  "features/#everyday",
		},
	}
}

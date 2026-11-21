package electricity

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "electricity",
			Shows: kernel.NewText("The electricity market price per hour for today and tomorrow, with the cheapest hours marked."),
			SetUp: kernel.NewText("Open the widget settings and choose your country. No account or key is needed."),
			Docs:  "modules/electricity/",
		},
	}
}

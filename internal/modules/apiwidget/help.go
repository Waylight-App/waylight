package apiwidget

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "apiwidget",
			Shows: kernel.NewText("Values from any web address that answers with JSON, such as a sensor or a counter."),
			SetUp: kernel.NewText("Open the widget settings, enter the address and pick the values to show."),
			Docs:  "modules/apiwidget/",
		},
	}
}

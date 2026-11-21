package focus

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "focus",
			Shows: kernel.NewText("A focus timer: work for a while, then take a short break."),
			SetUp: kernel.NewText("Press start. Open the widget settings to change the work and break lengths."),
			Docs:  "features/#everyday",
		},
	}
}

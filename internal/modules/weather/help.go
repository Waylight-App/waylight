package weather

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "weather",
			Shows: kernel.NewText("The forecast for a place: today, the next hours and days, and in Germany official weather warnings."),
			SetUp: kernel.NewText("Open the widget settings and pick a saved place or search for one. No account or key is needed."),
			Docs:  "features/#information",
		},
		{
			Kind:  "radar",
			Shows: kernel.NewText("A rain radar map of the last and next hours around a place in Germany."),
			SetUp: kernel.NewText("Open the widget settings and pick the place the map should centre on."),
			Docs:  "features/#information",
		},
	}
}

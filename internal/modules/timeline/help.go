package timeline

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "timeline",
			Shows: kernel.NewText("What happened when: outages, recoveries, failed backups and security events."),
			SetUp: kernel.NewText("Nothing to set up; events appear as the other widgets report them."),
			Docs:  "features/#home-server",
		},
	}
}

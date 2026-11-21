package backup

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "backup-health",
			Shows: kernel.NewText("A traffic light that says whether your newest backup is as recent as the schedule promises."),
			SetUp: kernel.NewText("Nothing to set up. Open Settings → Backups to schedule backups or run one now."),
			Docs:  "backups/",
		},
	}
}

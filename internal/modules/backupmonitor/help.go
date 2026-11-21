package backupmonitor

import (
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// WidgetHelp implements modules.Helpful (#678).
func (m *Module) WidgetHelp() []modules.WidgetHelp {
	return []modules.WidgetHelp{
		{
			Kind:  "backup-status",
			Shows: kernel.NewText("Whether your backups and scheduled jobs ran on time and without errors."),
			SetUp: kernel.NewText("In Settings under Backup monitoring, add a job and let it call its secret link after each run."),
			Docs:  "notifications/#backup-and-job-monitoring",
		},
	}
}

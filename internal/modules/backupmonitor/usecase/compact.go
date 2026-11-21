package usecase

import (
	"time"

	"waylight/internal/modules/backupmonitor/domain"
	"waylight/internal/shared/kernel"
)

// compactTimeLayout is the time of the last success in a compact state:
// numeric, so it reads the same in every language.
const compactTimeLayout = "2006-01-02 15:04"

// CompactState sums the backup sources up for the compact view (#594,
// #597): a failed run is a problem, an overdue one asks for attention,
// otherwise the latest success is shown. Times are in loc.
func CompactState(list []SourceView, now time.Time, loc *time.Location) kernel.CompactState {
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: now}
	if len(list) == 0 {
		state.Title = kernel.NewText("No backups yet")
		return state
	}
	var failed, overdue, waiting []string
	var last time.Time
	for _, s := range list {
		switch domain.State(s.State) {
		case domain.StateFailed:
			failed = append(failed, s.Name)
		case domain.StateOverdue:
			overdue = append(overdue, s.Name)
		case domain.StateWaiting:
			waiting = append(waiting, s.Name)
		}
		if s.LastSuccessAt.After(last) {
			last = s.LastSuccessAt
		}
	}
	total := len(list)
	switch {
	case len(failed) > 0:
		state.Level, state.Count = kernel.CompactProblem, kernel.IntPtr(len(failed))
		state.Title = kernel.NewTextf("{count} of {total} failed", "count", len(failed), "total", total)
		state.Detail = kernel.NewTextf("{name} failed", "name", failed[0])
	case len(overdue) > 0:
		state.Level, state.Count = kernel.CompactAttention, kernel.IntPtr(len(overdue))
		state.Title = kernel.NewTextf("{count} of {total} overdue", "count", len(overdue), "total", total)
		state.Detail = kernel.NewTextf("{name} is overdue", "name", overdue[0])
	case len(waiting) == total:
		state.Title = kernel.NewText("Waiting for the first run")
	default:
		state.Level = kernel.CompactOK
		state.Title = kernel.NewTextf("{total} backups OK", "total", total)
		if total == 1 {
			state.Title = kernel.NewText("Backup OK")
		}
	}
	if !last.IsZero() && state.Detail.Template == "" {
		state.Detail = kernel.NewTextf("Last success {time}", "time", last.In(loc).Format(compactTimeLayout))
	}
	return state
}

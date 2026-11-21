package usecase

import (
	"time"

	"waylight/internal/modules/diskhealth/domain"
	"waylight/internal/shared/kernel"
)

// CompactState sums the drives up for the compact view (#594, #597): a
// critical drive is a problem, a warning asks for attention, and drives
// without a recent report are not judged.
func CompactState(disks []domain.Disk, now time.Time) kernel.CompactState {
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: now}
	if len(disks) == 0 {
		state.Title = kernel.NewText("No drive reports yet")
		return state
	}
	var critical, warning []string
	healthy := 0
	for _, d := range disks {
		switch d.Status {
		case domain.StatusCritical:
			critical = append(critical, d.Device)
		case domain.StatusWarning:
			warning = append(warning, d.Device)
		case domain.StatusOK:
			healthy++
		}
	}
	total := len(disks)
	switch {
	case len(critical) > 0:
		state.Level, state.Count = kernel.CompactProblem, kernel.IntPtr(len(critical))
		state.Title = kernel.NewTextf("{count} of {total} drives failing", "count", len(critical), "total", total)
		state.Detail = kernel.NewTextf("{device} is failing", "device", critical[0])
	case len(warning) > 0:
		state.Level, state.Count = kernel.CompactAttention, kernel.IntPtr(len(warning))
		state.Title = kernel.NewTextf("{count} of {total} drives need a look", "count", len(warning), "total", total)
		state.Detail = kernel.NewTextf("{device} has warnings", "device", warning[0])
	case healthy == total:
		state.Level = kernel.CompactOK
		state.Title = kernel.NewTextf("{total} drives healthy", "total", total)
	default:
		state.Title = kernel.NewTextf("{healthy} of {total} drives healthy", "healthy", healthy, "total", total)
		state.Detail = kernel.NewText("Some reports are out of date")
	}
	return state
}

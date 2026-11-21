package usecase

import (
	"fmt"

	"waylight/internal/shared/kernel"
)

// CompactState shows the share of blocked queries for the compact view
// (#641). Blocking that is switched off asks for a look; a timed pause was
// started on purpose and only shows as detail.
func CompactState(v StatsView) kernel.CompactState {
	state := kernel.CompactState{
		Title:     kernel.NewTextf("{percent} % blocked", "percent", fmt.Sprintf("%.0f", v.BlockedPercent)),
		Level:     kernel.CompactNone,
		UpdatedAt: v.UpdatedAt,
	}
	switch {
	case v.Error != "":
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewText("The DNS filter is not reachable")
	case v.PausedUntil != nil:
		state.Detail = kernel.NewText("Blocking paused")
	case !v.ProtectionEnabled:
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewText("Blocking is off")
	default:
		state.Detail = kernel.NewTextf("{count} of {total} queries", "count", v.Blocked, "total", v.Queries)
	}
	return state
}

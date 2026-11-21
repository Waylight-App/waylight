package usecase

import "waylight/internal/shared/kernel"

// CompactState sums the container list up for the compact view (#638):
// how many run, and what needs a look. Stopped containers are often stopped
// on purpose, so only unhealthy ones and an unreachable Docker API colour it.
func CompactState(res ListResponse) kernel.CompactState {
	state := kernel.CompactState{
		Title:     kernel.NewTextf("{running} of {total} running", "running", res.Running, "total", res.Total),
		Level:     kernel.CompactNone,
		Count:     kernel.IntPtr(res.Unhealthy),
		UpdatedAt: res.UpdatedAt,
	}
	switch {
	case res.Error != "":
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewText("Docker is not reachable")
	case res.Unhealthy > 0:
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewTextf("{count} unhealthy", "count", res.Unhealthy)
	case res.Stopped > 0:
		state.Detail = kernel.NewTextf("{count} stopped", "count", res.Stopped)
	}
	return state
}

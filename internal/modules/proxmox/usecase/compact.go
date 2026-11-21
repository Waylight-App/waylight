package usecase

import "waylight/internal/shared/kernel"

// CompactState sums the cluster up for the compact view (#641): how many
// guests run. Stopped guests are often stopped on purpose, so only an
// offline node or an unreachable API asks for a look.
func CompactState(v StatusView) kernel.CompactState {
	running := 0
	for _, g := range v.Cluster.Guests {
		if g.Running() {
			running++
		}
	}
	offline := 0
	for _, n := range v.Cluster.Nodes {
		if !n.Online {
			offline++
		}
	}
	state := kernel.CompactState{
		Title: kernel.NewTextf("{running} of {total} guests running",
			"running", running, "total", len(v.Cluster.Guests)),
		Level:     kernel.CompactNone,
		Count:     kernel.IntPtr(offline),
		UpdatedAt: v.UpdatedAt,
	}
	switch {
	case v.Error != "":
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewText("Proxmox VE is not reachable")
	case offline > 0:
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewTextf("{offline} of {total} nodes offline", "offline", offline, "total", len(v.Cluster.Nodes))
	}
	return state
}

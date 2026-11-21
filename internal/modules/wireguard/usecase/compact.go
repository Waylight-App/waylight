package usecase

import "waylight/internal/shared/kernel"

// CompactState shows how many peers are connected for the compact view
// (#641). Peers come and go, so only an export that stopped updating asks
// for a look: the numbers would be outdated.
func CompactState(v ListView) kernel.CompactState {
	state := kernel.CompactState{
		Title: kernel.NewTextf("{online} of {total} peers connected",
			"online", v.Online, "total", len(v.Peers)),
		Level:     kernel.CompactNone,
		Count:     kernel.IntPtr(v.Online),
		UpdatedAt: v.UpdatedAt,
	}
	if v.Stale {
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewText("The export is out of date")
	}
	if len(v.Peers) == 0 {
		state.Title = kernel.NewText("No peers")
		state.Empty = !v.Stale
	}
	return state
}

package usecase

import (
	"fmt"

	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/shared/kernel"
)

// CompactState shows the latest measurement for the compact view (#638):
// download and upload in Mbit/s with the ping. A failed refresh asks for a
// look; the speed itself is judged by the alerts, not here.
func CompactState(v ResultsView) kernel.CompactState {
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: v.UpdatedAt}
	if v.Error != "" {
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewText("Speedtest Tracker is not reachable")
	}
	m := v.Latest
	if m == nil {
		state.Title = kernel.NewText("No results yet")
		return state
	}
	state.Title = kernel.NewTextf("{down} ↓ · {up} ↑ Mbit/s",
		"down", mbps(m.DownloadBps), "up", mbps(m.UploadBps))
	if v.Error == "" {
		state.Detail = kernel.NewTextf("Ping {ping} ms", "ping", fmt.Sprintf("%.0f", m.PingMs))
	}
	return state
}

// mbps rounds to whole Mbit/s; a compact view has no room for decimals.
func mbps(bps float64) string { return fmt.Sprintf("%.0f", domain.Mbps(bps)) }

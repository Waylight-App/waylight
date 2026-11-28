package usecase

import (
	"time"

	"waylight/internal/modules/services/domain"
	"waylight/internal/shared/kernel"
)

// CompactState sums the services up for the compact view (#594, #597):
// "11 of 12 running", with the service that is down (or the number of
// them) as detail. Tiles that lost their container and example tiles of
// the sample data (#854) are left out; services not checked yet count as
// neither up nor down.
func CompactState(list []ServiceView, now time.Time) kernel.CompactState {
	total, online, offline := 0, 0, 0
	var down string
	for _, s := range list {
		if s.Hidden || s.Example {
			continue
		}
		total++
		switch domain.State(s.State) {
		case domain.StateOnline:
			online++
		case domain.StateOffline:
			offline++
			if down == "" {
				down = s.Name
			}
		}
	}
	state := kernel.CompactState{Level: kernel.CompactNone, Count: kernel.IntPtr(offline), UpdatedAt: now}
	switch {
	case total == 0:
		state.Title = kernel.NewText("No services yet")
		state.Empty = true
		return state
	case offline == 0 && online == total:
		state.Level = kernel.CompactOK
	case offline > 0:
		state.Level = kernel.CompactProblem
	}
	state.Title = kernel.NewTextf("{online} of {total} running", "online", online, "total", total)
	switch {
	case offline == 1:
		state.Detail = kernel.NewTextf("{name} is down", "name", down)
	case offline > 1:
		state.Detail = kernel.NewTextf("{count} services down", "count", offline)
	}
	return state
}

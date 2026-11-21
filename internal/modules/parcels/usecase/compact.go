package usecase

import (
	"time"

	"waylight/internal/modules/parcels/domain"
	"waylight/internal/shared/kernel"
)

// CompactState sums the shown parcels up for the compact view (#638): how
// many are on the way and how many arrive today, in loc. A parcel the
// carrier reports a problem with asks for a look.
func CompactState(list []domain.Parcel, now time.Time, loc *time.Location) kernel.CompactState {
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: now}
	y, m, d := now.In(loc).Date()
	onTheWay, today, failed := 0, 0, 0
	for _, p := range list {
		if p.Delivered() {
			continue
		}
		onTheWay++
		if p.Status.Code == domain.StatusFailure {
			failed++
		}
		if !p.EstimatedDelivery.IsZero() {
			if ey, em, ed := p.EstimatedDelivery.In(loc).Date(); ey == y && em == m && ed == d {
				today++
			}
		}
	}
	state.Count = kernel.IntPtr(onTheWay)
	if onTheWay == 0 {
		state.Title = kernel.NewText("No parcels on the way")
		state.Empty = true
		return state
	}
	state.Title = kernel.NewTextf("{count} on the way", "count", onTheWay)
	switch {
	case failed > 0:
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewTextf("{count} with a problem", "count", failed)
	case today > 0:
		state.Detail = kernel.NewTextf("{count} arriving today", "count", today)
	}
	return state
}

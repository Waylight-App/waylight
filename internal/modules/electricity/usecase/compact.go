package usecase

import (
	"math"
	"strconv"
	"time"

	"waylight/internal/modules/electricity/domain"
	"waylight/internal/shared/kernel"
)

// compactClock is the time of day in compact texts.
const compactClock = "15:04"

// CompactState shows the price of the current hour and when the next of the
// cheapest hours begins (#641), e.g. "28 ct now" and "Cheap from 14:00", in
// loc. Prices that could not be refreshed ask for a look.
func CompactState(p Prices, now time.Time, loc *time.Location) kernel.CompactState {
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: p.FetchedAt}
	current := -1
	for i, h := range p.Hours {
		if !now.Before(h.Start) && now.Before(h.Start.Add(time.Hour)) {
			current = i
			break
		}
	}
	if current < 0 {
		state.Title = kernel.NewText("No price for this hour")
	} else {
		state.Title = kernel.NewTextf("{price} ct now", "price", cents(p.Hours[current].CentPerKWh))
		state.Detail = cheapDetail(p.Hours[current:], now, loc)
	}
	if p.Stale {
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewText("Prices could not be updated")
	}
	return state
}

// cheapDetail names the next cheap hour, starting with the current one;
// zero when none is left in the published prices.
func cheapDetail(hours []domain.Hour, now time.Time, loc *time.Location) kernel.Text {
	for i, h := range hours {
		if !h.Cheapest {
			continue
		}
		if i == 0 {
			return kernel.NewText("Cheap now")
		}
		start := h.Start.In(loc)
		if y, m, d := now.In(loc).Date(); start.Year() == y && start.Month() == m && start.Day() == d {
			return kernel.NewTextf("Cheap from {time}", "time", start.Format(compactClock))
		}
		return kernel.NewTextf("Cheap tomorrow from {time}", "time", start.Format(compactClock))
	}
	return kernel.Text{}
}

// cents rounds to whole cents; a compact view has no room for decimals, and
// whole numbers need no locale-specific separator.
func cents(v float64) string { return strconv.Itoa(int(math.Round(v))) }

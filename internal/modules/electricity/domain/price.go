// Package domain holds the electricity price rules (#288).
//
// Ring: Entities.
package domain

import (
	"slices"
	"time"

	"waylight/internal/shared/kernel"
)

// Zone is a bidding zone of the day-ahead market.
type Zone string

// Supported zones; providers without an API key cover these.
const (
	ZoneDELU Zone = "DE-LU" // Germany and Luxembourg
	ZoneAT   Zone = "AT"    // Austria
)

// Zones lists the supported zones in display order.
var Zones = []Zone{ZoneDELU, ZoneAT}

// ErrUnknownZone rejects zones no provider serves.
var ErrUnknownZone = kernel.NewValidationError("zone", "Choose a supported bidding zone (DE-LU or AT).")

// ParseZone validates a zone; empty means DE-LU.
func ParseZone(s string) (Zone, error) {
	if s == "" {
		return ZoneDELU, nil
	}
	if !slices.Contains(Zones, Zone(s)) {
		return "", ErrUnknownZone
	}
	return Zone(s), nil
}

// Slot is one market price interval (15 or 60 minutes).
type Slot struct {
	Start, End time.Time
	EurPerMWh  float64
}

// Hour is the average price of one hour, in cent per kWh.
type Hour struct {
	Start      time.Time
	CentPerKWh float64
	Cheapest   bool
}

// Hourly averages the slots per hour (the day-ahead market has used 15-minute
// intervals since October 2025; an hourly view is easier to read) and marks
// the cheapest hours of each day. Hours are in loc, sorted.
func Hourly(slots []Slot, loc *time.Location, cheapestPerDay int) []Hour {
	type acc struct {
		sum float64
		n   int
	}
	sums := map[time.Time]*acc{}
	for _, s := range slots {
		h := s.Start.In(loc).Truncate(time.Hour)
		a := sums[h]
		if a == nil {
			a = &acc{}
			sums[h] = a
		}
		a.sum += s.EurPerMWh
		a.n++
	}
	hours := make([]Hour, 0, len(sums))
	for start, a := range sums {
		// €/MWh to ct/kWh: 1 €/MWh = 0.1 ct/kWh.
		hours = append(hours, Hour{Start: start, CentPerKWh: round2(a.sum / float64(a.n) / 10)})
	}
	slices.SortFunc(hours, func(a, b Hour) int { return a.Start.Compare(b.Start) })
	markCheapest(hours, loc, cheapestPerDay)
	return hours
}

func markCheapest(hours []Hour, loc *time.Location, n int) {
	byDay := map[string][]int{}
	for i, h := range hours {
		d := h.Start.In(loc).Format(time.DateOnly)
		byDay[d] = append(byDay[d], i)
	}
	for _, idx := range byDay {
		sorted := slices.Clone(idx)
		slices.SortStableFunc(sorted, func(a, b int) int {
			switch {
			case hours[a].CentPerKWh < hours[b].CentPerKWh:
				return -1
			case hours[a].CentPerKWh > hours[b].CentPerKWh:
				return 1
			}
			return 0
		})
		for _, i := range sorted[:min(n, len(sorted))] {
			hours[i].Cheapest = true
		}
	}
}

func round2(v float64) float64 {
	if v < 0 {
		return -round2(-v)
	}
	return float64(int64(v*100+0.5)) / 100
}

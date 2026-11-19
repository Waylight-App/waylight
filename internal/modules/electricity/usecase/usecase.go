// Package usecase holds the electricity price use cases (#288).
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"waylight/internal/modules/electricity/domain"
	"waylight/internal/shared/kernel"
)

// cacheTTL: day-ahead prices change once a day (tomorrow's are published
// around 13:00), so a quarter of an hour is plenty (cache-aside, ADR 0006).
const cacheTTL = 15 * time.Minute

// CheapestPerDay is how many hours of each day are marked as cheapest.
const CheapestPerDay = 3

// Prices is the answer of the widget.
type Prices struct {
	Zone      domain.Zone
	Hours     []domain.Hour
	FetchedAt time.Time
	// Stale is set when the provider failed and the last known prices are shown.
	Stale bool
}

// PriceProvider reads day-ahead prices (output port).
type PriceProvider interface {
	Slots(ctx context.Context, zone domain.Zone, from, to time.Time) ([]domain.Slot, error)
}

// Electricity is the input port of the module.
type Electricity interface {
	// Prices returns today's and, once published, tomorrow's hourly prices.
	Prices(ctx context.Context, zone string) (Prices, error)
}

type entry struct {
	prices  Prices
	fetched time.Time
}

type electricity struct {
	provider PriceProvider
	clock    kernel.Clock
	loc      *time.Location
	mu       sync.Mutex
	cache    map[domain.Zone]entry
}

// New returns the use cases; loc is the time zone of "today" (nil = local).
func New(provider PriceProvider, clock kernel.Clock, loc *time.Location) Electricity {
	if loc == nil {
		loc = time.Local
	}
	return &electricity{provider: provider, clock: clock, loc: loc, cache: map[domain.Zone]entry{}}
}

func (uc *electricity) Prices(ctx context.Context, zone string) (Prices, error) {
	z, err := domain.ParseZone(zone)
	if err != nil {
		return Prices{}, err
	}
	now := uc.clock.Now()
	uc.mu.Lock()
	cached, ok := uc.cache[z]
	uc.mu.Unlock()
	today := time.Date(now.In(uc.loc).Year(), now.In(uc.loc).Month(), now.In(uc.loc).Day(), 0, 0, 0, 0, uc.loc)
	if ok && now.Sub(cached.fetched) < cacheTTL && !cached.prices.Hours[0].Start.Before(today) {
		return cached.prices, nil
	}
	slots, err := uc.provider.Slots(ctx, z, today, today.AddDate(0, 0, 2))
	if err != nil {
		if ok {
			if stale, ok := fromDay(cached.prices, today); ok {
				return stale, nil
			}
		}
		var pe *kernel.PublicError
		if errors.As(err, &pe) {
			return Prices{}, err
		}
		return Prices{}, kernel.TargetFailure(kernel.ErrUnavailable, fmt.Errorf("electricity prices: %w", err))
	}
	hours := domain.Hourly(slots, uc.loc, CheapestPerDay)
	if len(hours) == 0 {
		return Prices{}, kernel.NewPublicError(kernel.ErrUnavailable, "No electricity prices were published for today yet.")
	}
	p := Prices{Zone: z, Hours: hours, FetchedAt: now}
	uc.mu.Lock()
	uc.cache[z] = entry{prices: p, fetched: now}
	uc.mu.Unlock()
	return p, nil
}

// fromDay returns the cached prices without the hours before day, marked
// stale; false when none are left. After midnight a failed refresh must not
// show yesterday's prices as today's.
func fromDay(p Prices, day time.Time) (Prices, bool) {
	i := 0
	for i < len(p.Hours) && p.Hours[i].Start.Before(day) {
		i++
	}
	if i == len(p.Hours) {
		return Prices{}, false
	}
	p.Hours = p.Hours[i:]
	p.Stale = true
	return p, true
}

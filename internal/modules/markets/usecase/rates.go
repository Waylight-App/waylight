package usecase

import (
	"context"
	"sync"
	"time"

	"trisse-home/internal/shared/kernel"
)

// RateSource returns the newest euro reference rates (output port, the ECB).
type RateSource interface {
	Latest(ctx context.Context) (day time.Time, rates map[string]float64, err error)
}

// RatesView is the answer for currency conversion in the search (#482):
// how many units of each currency one euro buys on Day.
type RatesView struct {
	Base  string
	Day   time.Time
	Rates map[string]float64
	// Stale is set when the source could not be reached and the last rates
	// that were fetched are returned instead.
	Stale bool
}

// Errors.
var (
	ErrRatesOff         = kernel.NewPublicError(kernel.ErrNotFound, "Currency rates are switched off on this server.")
	ErrRatesUnavailable = kernel.NewPublicError(kernel.ErrUnavailable, "The currency rates could not be fetched; try again later.")
)

// rateCacheFor is how long fetched rates are used before asking again; the
// ECB publishes once per working day.
const rateCacheFor = 24 * time.Hour

// retryAfter keeps a failing source from being asked on every keystroke.
const retryAfter = 15 * time.Minute

// Rates fetches and caches the reference rates.
type Rates struct {
	source RateSource
	clock  kernel.Clock
	on     bool

	mu      sync.Mutex
	cached  *RatesView
	fetched time.Time
	failed  time.Time
}

// NewRates returns the use case; on false switches currency rates off.
func NewRates(source RateSource, clock kernel.Clock, on bool) *Rates {
	return &Rates{source: source, clock: clock, on: on}
}

// Current returns the rates, fetching them at most once a day; when the
// source fails it falls back to the last rates it has, marked stale.
func (r *Rates) Current(ctx context.Context) (RatesView, error) {
	if !r.on {
		return RatesView{}, ErrRatesOff
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.clock.Now()
	if r.cached != nil && now.Sub(r.fetched) < rateCacheFor {
		return *r.cached, nil
	}
	if !r.failed.IsZero() && now.Sub(r.failed) < retryAfter {
		return r.fallback()
	}
	day, rates, err := r.source.Latest(ctx)
	if err != nil {
		r.failed = now
		return r.fallback()
	}
	rates["EUR"] = 1
	r.cached = &RatesView{Base: "EUR", Day: day, Rates: rates}
	r.fetched, r.failed = now, time.Time{}
	return *r.cached, nil
}

func (r *Rates) fallback() (RatesView, error) {
	if r.cached == nil {
		return RatesView{}, ErrRatesUnavailable
	}
	v := *r.cached
	v.Stale = true
	return v, nil
}

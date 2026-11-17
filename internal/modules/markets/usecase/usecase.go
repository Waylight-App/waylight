// Package usecase holds the market watchlist use cases (#284).
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"sync"
	"time"

	"waylight/internal/modules/markets/domain"
	"waylight/internal/shared/kernel"
)

// cacheTTL: daily closes and reference rates change once a day; an hour
// keeps the sources' load low (cache-aside, ADR 0006).
const cacheTTL = time.Hour

// HistoryDays is how far back the sparkline goes.
const HistoryDays = 30

// SeriesProvider reads daily closes of one source (output port).
type SeriesProvider interface {
	Series(ctx context.Context, sym domain.Symbol, from, to time.Time) ([]domain.Point, error)
}

// Row errors are codes the UI translates.
const (
	ErrNoData      = "no_data"     // the source knows no such symbol
	ErrUnavailable = "unavailable" // the source did not answer
)

// Row is one watchlist entry: a quote, or why there is none.
type Row struct {
	Symbol domain.Symbol
	Quote  *domain.Quote
	Stale  bool
	Error  string
}

// Markets is the input port.
type Markets interface {
	Watchlist(ctx context.Context, symbols []string) ([]Row, error)
}

type entry struct {
	at    time.Time
	quote domain.Quote
}

type markets struct {
	providers map[domain.Source]SeriesProvider
	clock     kernel.Clock
	mu        sync.Mutex
	cache     map[domain.Symbol]entry
}

// New returns the use cases with one provider per source.
func New(providers map[domain.Source]SeriesProvider, clock kernel.Clock) Markets {
	return &markets{providers: providers, clock: clock, cache: map[domain.Symbol]entry{}}
}

func (uc *markets) Watchlist(ctx context.Context, symbols []string) ([]Row, error) {
	syms, err := domain.ParseSymbols(symbols)
	if err != nil {
		return nil, err
	}
	now := uc.clock.Now()
	rows := make([]Row, len(syms))
	var wg sync.WaitGroup
	for i, sym := range syms {
		wg.Go(func() { rows[i] = uc.row(ctx, sym, now) })
	}
	wg.Wait()
	return rows, nil
}

// row returns one entry; a failing source shows the last known quote.
func (uc *markets) row(ctx context.Context, sym domain.Symbol, now time.Time) Row {
	uc.mu.Lock()
	c, ok := uc.cache[sym]
	uc.mu.Unlock()
	if ok && now.Sub(c.at) < cacheTTL {
		q := c.quote
		return Row{Symbol: sym, Quote: &q}
	}
	p := uc.providers[sym.Source]
	series, err := p.Series(ctx, sym, now.AddDate(0, 0, -HistoryDays-7), now)
	q, found := domain.NewQuote(sym, lastDays(series, now))
	switch {
	case err == nil && found:
		uc.mu.Lock()
		uc.cache[sym] = entry{at: now, quote: q}
		uc.mu.Unlock()
		return Row{Symbol: sym, Quote: &q}
	case ok:
		old := c.quote
		return Row{Symbol: sym, Quote: &old, Stale: true}
	case err == nil:
		return Row{Symbol: sym, Error: ErrNoData}
	}
	return Row{Symbol: sym, Error: ErrUnavailable}
}

func lastDays(series []domain.Point, now time.Time) []domain.Point {
	from := now.AddDate(0, 0, -HistoryDays)
	for i, p := range series {
		if !p.Day.Before(from) {
			return series[i:]
		}
	}
	return series
}

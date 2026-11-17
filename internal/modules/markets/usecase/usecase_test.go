package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/markets/domain"
	"waylight/internal/shared/kernel"
)

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type fake struct {
	calls int
	fail  bool
	empty bool
}

func (f *fake) Series(_ context.Context, _ domain.Symbol, _, to time.Time) ([]domain.Point, error) {
	f.calls++
	if f.fail {
		return nil, errors.New("timeout")
	}
	if f.empty {
		return nil, nil
	}
	return []domain.Point{{Day: to.AddDate(0, 0, -60), Value: 1}, {Day: to.AddDate(0, 0, -2), Value: 100}, {Day: to.AddDate(0, 0, -1), Value: 110}}, nil
}

func TestWatchlist(t *testing.T) {
	ctx := context.Background()
	c := &clk{t: time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)}
	fx, stooq := &fake{}, &fake{}
	uc := New(map[domain.Source]SeriesProvider{domain.SourceFX: fx, domain.SourceStooq: stooq}, c)
	rows, err := uc.Watchlist(ctx, []string{"fx:USD", "stooq:aapl.us"})
	if err != nil || len(rows) != 2 || rows[1].Quote.Value != 110 || rows[1].Quote.ChangePct != 10 || len(rows[1].Quote.History) != 2 {
		t.Fatalf("history trimmed to 30 days: %+v %v", rows, err)
	}
	_, _ = uc.Watchlist(ctx, []string{"fx:USD"})
	if fx.calls != 1 {
		t.Fatal("not cached")
	}
	c.t = c.t.Add(2 * time.Hour)
	fx.fail = true
	if rows, _ := uc.Watchlist(ctx, []string{"fx:USD"}); !rows[0].Stale || rows[0].Quote == nil {
		t.Fatalf("stale: %+v", rows[0])
	}
	stooq.empty = true
	if rows, _ := uc.Watchlist(ctx, []string{"stooq:nope.us"}); rows[0].Error == "" || rows[0].Quote != nil {
		t.Fatalf("no data: %+v", rows[0])
	}
	if _, err := uc.Watchlist(ctx, []string{"bitcoin"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
}

package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/electricity/domain"
	"trisse-home/internal/shared/kernel"
)

type clk struct{ t time.Time }

func (c *clk) Now() time.Time { return c.t }

type fakeProvider struct {
	calls int
	fail  bool
	from  time.Time
}

func (f *fakeProvider) Slots(_ context.Context, _ domain.Zone, from, _ time.Time) ([]domain.Slot, error) {
	f.calls++
	f.from = from
	if f.fail {
		return nil, errors.New("connection refused")
	}
	return []domain.Slot{{Start: from, EurPerMWh: 100}, {Start: from.Add(time.Hour), EurPerMWh: 50}}, nil
}

func TestPrices(t *testing.T) {
	ctx := context.Background()
	loc := time.FixedZone("CET", 3600)
	c := &clk{t: time.Date(2026, 11, 14, 9, 30, 0, 0, loc)}
	p := &fakeProvider{}
	uc := New(p, c, loc)
	got, err := uc.Prices(ctx, "")
	if err != nil || got.Zone != domain.ZoneDELU || len(got.Hours) != 2 || got.Hours[1].CentPerKWh != 5 || !got.Hours[1].Cheapest {
		t.Fatalf("%+v %v", got, err)
	}
	if !p.from.Equal(time.Date(2026, 11, 14, 0, 0, 0, 0, loc)) {
		t.Fatal("must ask from midnight of today:", p.from)
	}
	// Cached for 15 minutes.
	c.t = c.t.Add(10 * time.Minute)
	_, _ = uc.Prices(ctx, "DE-LU")
	if p.calls != 1 {
		t.Fatal("not cached")
	}
	// A failing provider shows the last known prices, marked stale.
	c.t = c.t.Add(10 * time.Minute)
	p.fail = true
	if got, err := uc.Prices(ctx, "DE-LU"); err != nil || !got.Stale {
		t.Fatalf("stale: %+v %v", got, err)
	}
	// Without anything cached the error is shown.
	if _, err := uc.Prices(ctx, "AT"); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal(err)
	}
	if _, err := uc.Prices(ctx, "FR"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
}

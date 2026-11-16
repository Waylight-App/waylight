package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRates struct {
	calls int
	fail  bool
}

func (f *fakeRates) Latest(context.Context) (time.Time, map[string]float64, error) {
	f.calls++
	if f.fail {
		return time.Time{}, nil, errors.New("ecb down")
	}
	return time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC), map[string]float64{"USD": 1.08, "CHF": 0.94}, nil
}

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

func TestRatesAreCachedForADayAndFallBack(t *testing.T) {
	src, clk := &fakeRates{}, &movingClock{t: time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)}
	r := NewRates(src, clk, true)
	v, err := r.Current(context.Background())
	if err != nil || v.Rates["USD"] != 1.08 || v.Rates["EUR"] != 1 || v.Base != "EUR" || v.Stale {
		t.Fatalf("%+v %v", v, err)
	}
	clk.t = clk.t.Add(23 * time.Hour)
	_, _ = r.Current(context.Background())
	if src.calls != 1 {
		t.Fatalf("fetched %d times within a day", src.calls)
	}
	// A day later the source is down: the last rates, marked stale.
	clk.t = clk.t.Add(2 * time.Hour)
	src.fail = true
	if v, err := r.Current(context.Background()); err != nil || !v.Stale || v.Rates["USD"] != 1.08 {
		t.Fatalf("fallback: %+v %v", v, err)
	}
	// It is not asked again on every keystroke.
	_, _ = r.Current(context.Background())
	if src.calls != 2 {
		t.Fatalf("source asked %d times while failing", src.calls)
	}
	clk.t = clk.t.Add(16 * time.Minute)
	src.fail = false
	if v, _ := r.Current(context.Background()); v.Stale || src.calls != 3 {
		t.Fatalf("after recovery: %+v, %d calls", v, src.calls)
	}
}

func TestRatesWithoutCacheOrSwitchedOff(t *testing.T) {
	if _, err := NewRates(&fakeRates{fail: true}, &movingClock{t: time.Now()}, true).Current(context.Background()); !errors.Is(err, ErrRatesUnavailable) {
		t.Fatalf("first fetch failing: %v", err)
	}
	src := &fakeRates{}
	if _, err := NewRates(src, &movingClock{t: time.Now()}, false).Current(context.Background()); !errors.Is(err, ErrRatesOff) || src.calls != 0 {
		t.Fatalf("off: %v, %d calls", err, src.calls)
	}
}

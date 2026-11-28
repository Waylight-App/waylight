package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/weather/domain"
)

type fakeProvider struct {
	calls int
	err   error
}

func (f *fakeProvider) Forecast(_ context.Context, loc domain.Location) (domain.Forecast, error) {
	f.calls++
	if f.err != nil {
		return domain.Forecast{}, f.err
	}
	return domain.Forecast{Current: domain.Current{Temperature: float64(f.calls)}}, nil
}

func (f *fakeProvider) SearchPlaces(context.Context, string) ([]domain.Place, error) {
	return []domain.Place{{Name: "Oldenburg", Country: "Germany", Lat: 53.14, Lon: 8.21}}, nil
}

// fakeCache is a minimal in-memory ForecastCache for tests.
type fakeCache struct {
	forecasts map[string]domain.Forecast
	seen      map[string]time.Time
	locs      map[string]domain.Location
}

func newFakeCache() *fakeCache {
	return &fakeCache{forecasts: map[string]domain.Forecast{}, seen: map[string]time.Time{}, locs: map[string]domain.Location{}}
}

func (c *fakeCache) Get(k string) (domain.Forecast, bool) { f, ok := c.forecasts[k]; return f, ok }
func (c *fakeCache) Put(f domain.Forecast)                { c.forecasts[f.Location.Key()] = f }
func (c *fakeCache) Touch(l domain.Location, at time.Time) {
	c.seen[l.Key()], c.locs[l.Key()] = at, l
}

func (c *fakeCache) Locations(since time.Time) []domain.Location {
	var out []domain.Location
	for k, at := range c.seen {
		if at.Before(since) {
			delete(c.seen, k)
			continue
		}
		out = append(out, c.locs[k])
	}
	return out
}

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

func TestForecastCacheAside(t *testing.T) {
	ctx := context.Background()
	p := &fakeProvider{}
	c := &clock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	uc := NewWeather(p, newFakeCache(), c)

	first, err := uc.GetForecast(ctx, ForecastRequest{Lat: 53.14, Lon: 8.21})
	if err != nil || first.Temperature != 1 {
		t.Fatalf("first fetch: %+v %v", first, err)
	}
	if again, _ := uc.GetForecast(ctx, ForecastRequest{Lat: 53.141, Lon: 8.211}); again.Temperature != 1 || p.calls != 1 {
		t.Fatal("cached value not used for nearby coordinates")
	}
	c.now = c.now.Add(MaxForecastAge)
	if fresh, _ := uc.GetForecast(ctx, ForecastRequest{Lat: 53.14, Lon: 8.21}); fresh.Temperature != 2 {
		t.Fatal("stale value not refreshed")
	}
	p.err = errors.New("down")
	c.now = c.now.Add(MaxForecastAge)
	if stale, err := uc.GetForecast(ctx, ForecastRequest{Lat: 53.14, Lon: 8.21}); err != nil || stale.Temperature != 2 {
		t.Fatalf("stale fallback failed: %v", err)
	}
	if _, err := uc.GetForecast(ctx, ForecastRequest{Lat: 1, Lon: 1}); err == nil {
		t.Fatal("expected error without cached value")
	}
	if _, err := uc.GetForecast(ctx, ForecastRequest{Lat: 100, Lon: 1}); err == nil {
		t.Fatal("expected validation error")
	}
}

// sunProvider reports a day flag from the fetch and the sun times of the day.
type sunProvider struct{ fakeProvider }

func (s *sunProvider) Forecast(ctx context.Context, loc domain.Location) (domain.Forecast, error) {
	f, err := s.fakeProvider.Forecast(ctx, loc)
	f.Current.IsDay = true
	f.Days = []domain.Day{{
		Sunrise: time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC), Sunset: time.Date(2026, 1, 1, 16, 0, 0, 0, time.UTC),
	}}
	return f, err
}

// The tile (compact state), the panel and the top bar (forecast) read one
// observation; its day/night state follows the sunset, not the fetch (#853).
func TestOneObservationForEveryDisplay(t *testing.T) {
	ctx := context.Background()
	c := &clock{now: time.Date(2026, 1, 1, 15, 50, 0, 0, time.UTC)}
	uc := NewWeather(&sunProvider{}, newFakeCache(), c)
	panel, _ := uc.GetForecast(ctx, ForecastRequest{Lat: 53.14, Lon: 8.21})
	if !panel.IsDay {
		t.Fatal("before sunset it is day")
	}
	c.now = c.now.Add(15 * time.Minute)
	panel, _ = uc.GetForecast(ctx, ForecastRequest{Lat: 53.14, Lon: 8.21})
	tile := CompactState(panel)
	if panel.IsDay || panel.Temperature != 1 || tile.Title.Params["temp"] != "1" {
		t.Fatalf("after sunset the cached observation must read as night: %+v, tile %+v", panel, tile)
	}
	if panel.Days[0].Sunset.IsZero() {
		t.Fatal("the sun times must reach the clients")
	}
}

func TestRefreshAllForgetsUnusedLocations(t *testing.T) {
	ctx := context.Background()
	p := &fakeProvider{}
	c := &clock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	uc := NewWeather(p, newFakeCache(), c)
	_, _ = uc.GetForecast(ctx, ForecastRequest{Lat: 1, Lon: 1})
	if err := uc.RefreshAll(ctx); err != nil || p.calls != 2 {
		t.Fatalf("refresh: calls=%d err=%v", p.calls, err)
	}
	c.now = c.now.Add(UnusedLocationTTL + time.Minute)
	_ = uc.RefreshAll(ctx)
	if p.calls != 2 {
		t.Fatal("unused location still refreshed")
	}
}

func TestSearchPlaces(t *testing.T) {
	uc := NewWeather(&fakeProvider{}, newFakeCache(), &clock{})
	if _, err := uc.SearchPlaces(context.Background(), "x"); err == nil {
		t.Fatal("expected validation error")
	}
	got, err := uc.SearchPlaces(context.Background(), "Oldenburg")
	if err != nil || len(got) != 1 || got[0].Country != "Germany" {
		t.Fatalf("unexpected %+v %v", got, err)
	}
}

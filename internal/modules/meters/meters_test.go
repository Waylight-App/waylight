package meters

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/meters/adapter/gateway"
	"waylight/internal/modules/meters/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return strconv.Itoa(s.n) }

func newModule(t *testing.T) *Module {
	t.Helper()
	return New(dbtest.Open(t, "meters", gateway.Migrations()), fixedClock{time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)}, &seqIDs{}, time.UTC)
}

func TestCompactState(t *testing.T) {
	m := newModule(t)
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if s, _ := m.CompactState(ctx); !s.Empty || s.Title.String() != "No meters yet" {
		t.Fatalf("empty: %+v", s)
	}
	power, _ := m.uc.Create(ctx, domain.Input{Name: "Power", Kind: "electricity", Unit: "kWh"}, "")
	water, _ := m.uc.Create(ctx, domain.Input{Name: "Water", Kind: "water", Unit: "m³"}, "")
	if s, _ := m.CompactState(ctx); !s.Empty || s.Title.String() != "No readings yet" {
		t.Fatalf("no readings: %+v", s)
	}
	for id, day := range map[string]string{power.ID: "2026-11-22", water.ID: "2026-11-12"} {
		if _, err := m.uc.AddReading(ctx, id, domain.ReadingInput{Day: day, Value: 5}); err != nil {
			t.Fatal(err)
		}
	}
	if s, err := m.CompactState(ctx); err != nil || s.Title.String() != "Water: read 10 days ago" || s.Empty {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestReadingRules(t *testing.T) {
	m := newModule(t)
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	v, err := m.uc.Create(anna, domain.Input{Name: "Gas", Kind: "gas", Unit: "m³"}, "")
	if err != nil || v.Access.Visibility != kernel.Everyone {
		t.Fatalf("a new meter is shared with everyone: %+v %v", v, err)
	}
	for _, c := range []struct {
		who context.Context
		in  domain.ReadingInput
		ok  bool
	}{
		{anna, domain.ReadingInput{Day: "2026-10-01", Value: 100}, true},
		{ben, domain.ReadingInput{Day: "2026-11-01", Value: 130}, true},
		{ben, domain.ReadingInput{Day: "2026-10-15", Value: 90}, false},
		{ben, domain.ReadingInput{Day: "2026-10-15", Value: 140}, false},
	} {
		if _, err := m.uc.AddReading(c.who, v.ID, c.in); (err == nil) != c.ok {
			t.Errorf("%+v: %v", c.in, err)
		}
	}
	if err := m.uc.DeleteReading(ben, v.ID, "2026-11-01"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben removed a reading: %v", err)
	}
	d, err := m.uc.Detail(ben, v.ID, 0)
	if err != nil || len(d.Readings) != 2 || d.Readings[0].Value != 130 || d.Last.Value != 130 || d.Consumption.Year != 2026 || d.Mine {
		t.Fatalf("%+v %v", d, err)
	}
	if err := m.uc.DeleteReading(anna, v.ID, "2026-11-01"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.uc.Update(ben, v.ID, domain.Input{Name: "x", Kind: "gas", Unit: "m³"}, ""); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed the meter: %v", err)
	}
	private, _ := m.uc.Update(anna, v.ID, domain.Input{Name: "Gas", Kind: "gas", Unit: "m³"}, "private")
	if _, err := m.uc.AddReading(ben, private.ID, domain.ReadingInput{Day: "2026-11-02", Value: 150}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben added to a private meter: %v", err)
	}
}

func TestSeedDemo(t *testing.T) {
	m := newModule(t)
	ctx := context.Background()
	for range 2 { // the second run must not add anything
		if err := m.SeedDemo(ctx); err != nil {
			t.Fatal(err)
		}
	}
	admin := kernel.WithPrincipal(ctx, kernel.AdminPrincipal())
	list, _ := m.uc.List(admin)
	if len(list) != 2 || list[0].Last == nil {
		t.Fatalf("%d meters", len(list))
	}
	d, err := m.uc.Detail(admin, list[0].ID, 2026)
	if err != nil || len(d.Readings) != 13 || !d.Consumption.Months[9].Known {
		t.Fatalf("%+v %v", len(d.Readings), err)
	}
}

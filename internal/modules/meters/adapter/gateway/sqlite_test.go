package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/meters/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

func meter(t *testing.T, id string, v kernel.Visibility) domain.Meter {
	t.Helper()
	m, err := domain.New(id, domain.Input{Name: "Meter " + id, Kind: "water", Unit: "m³"},
		kernel.Access{Visibility: v}, time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func reading(s string, v float64) domain.Reading {
	d, _ := domain.ParseDay("day", s)
	return domain.Reading{Day: d, Value: v}
}

// TestOwnershipAndReadings proves ADR 0018 for meters: everyone who sees a
// meter adds readings, only the owner changes or removes it and its readings.
func TestOwnershipAndReadings(t *testing.T) {
	db := dbtest.Open(t, "meters", Migrations())
	repo := NewRepository(db)
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	shared, private := meter(t, "shared", kernel.Everyone), meter(t, "private", kernel.Private)
	for _, m := range []domain.Meter{shared, private} {
		if err := repo.Save(anna, m); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.SaveReading(anna, "private", reading("2026-01-01", 5)); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveReading(anna, "shared", reading("2026-01-01", 10)); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveReading(ben, "shared", reading("2026-02-01", 20)); err != nil {
		t.Fatalf("ben adds to a shared meter: %v", err)
	}
	if err := repo.SaveReading(ben, "shared", reading("2026-02-01", 21)); err != nil { // same day replaces
		t.Fatal(err)
	}
	if err := repo.SaveReading(ben, "private", reading("2026-02-01", 9)); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben added to anna's private meter: %v", err)
	}
	list, _ := repo.List(ben)
	last, _ := repo.LastReadings(ben)
	if len(list) != 1 || list[0].ID != "shared" || len(last) != 1 || last["shared"].Value != 21 {
		t.Fatalf("ben sees %+v %+v", list, last)
	}
	if rs, err := repo.Readings(ben, "shared"); err != nil || len(rs) != 2 || rs[0].Value != 10 || rs[1].Value != 21 {
		t.Fatalf("%+v %v", rs, err)
	}
	if rs, _ := repo.Readings(ben, "private"); len(rs) != 0 {
		t.Fatalf("ben reads anna's private readings: %+v", rs)
	}
	shared.Name = "Mine"
	if err := repo.Save(ben, shared); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's meter: %v", err)
	}
	if err := repo.DeleteReading(ben, "shared", reading("2026-02-01", 0).Day); !errors.Is(err, domain.ErrNoReading) {
		t.Fatalf("ben removed a reading: %v", err)
	}
	if err := repo.Delete(ben, "shared"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted anna's meter: %v", err)
	}
	if err := repo.DeleteReading(anna, "shared", reading("2026-02-01", 0).Day); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(anna, "shared"); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM meters_readings`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("readings left: %d %v", n, err)
	}
}

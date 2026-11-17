package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/countdown/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestCountdownOwnership proves ADR 0018 for countdowns.
func TestCountdownOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "countdown", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})

	trip, _ := domain.New("c-anna", domain.Input{Title: "Surprise party", Date: "2026-12-24"}, now)
	holiday, _ := domain.New("c-admin", domain.Input{Title: "Holiday", Date: "2026-12-20"}, now)
	if err := repo.Save(anna, trip); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(admin, holiday); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(ben); len(list) != 1 || list[0].ID != "c-admin" {
		t.Fatalf("ben sees %v", list)
	}
	if _, err := repo.Get(admin, "c-anna"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the admin read anna's countdown: %v", err)
	}
	holiday.Title = "Mine"
	if err := repo.Save(ben, holiday); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed the admin's countdown: %v", err)
	}
	if err := repo.Delete(ben, "c-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's countdown: %v", err)
	}
	if got, _ := repo.Get(admin, "c-admin"); got.Title != "Holiday" {
		t.Fatalf("title changed to %q", got.Title)
	}
	trip.Title = "Party"
	if err := repo.Save(anna, trip); err != nil {
		t.Fatal("the owner may change her countdown:", err)
	}
}

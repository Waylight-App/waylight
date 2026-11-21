package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/places/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

func TestRepository(t *testing.T) {
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	repo := NewRepository(dbtest.Open(t, "places", Migrations()))
	p, _ := domain.NewPlace("1", "Home", "Oldenburg, Germany", 53.14, 8.21, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	if err := repo.Save(ctx, p); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, "1")
	if err != nil || got != p {
		t.Fatalf("%+v %v", got, err)
	}
	_ = repo.Delete(ctx, "1")
	if _, err := repo.Get(ctx, "1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

// TestOwnership proves ADR 0018 for places: others may see a shared place
// but neither change nor delete it.
func TestOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "places", Migrations()))
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})

	home, _ := domain.NewPlace("p-admin", "Home", "", 53, 8, now)
	if err := repo.Save(admin, home); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(ben); len(list) != 1 {
		t.Fatalf("ben sees %v", list)
	}
	home.Name = "Mine"
	if err := repo.Save(ben, home); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben renamed the admin's place: %v", err)
	}
	if err := repo.Delete(ben, "p-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's place: %v", err)
	}
}

package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/habits/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestHabitOwnership proves ADR 0018 for habits, including ticking off.
func TestHabitOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "habits", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	day := time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	every, _ := domain.NewSchedule([]int{0, 1, 2, 3, 4, 5, 6})

	run, _ := domain.New("h-anna", "Run", "🏃", every, 0, now)
	read, _ := domain.New("h-admin", "Read", "📖", every, 1, now)
	if err := repo.Save(anna, run); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(admin, read); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetCompletion(anna, "h-anna", day, true); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(ben); len(list) != 1 || list[0].ID != "h-admin" {
		t.Fatalf("ben sees %v", list)
	}
	if done, _ := repo.Completions(ben, day); len(done["h-anna"]) != 0 {
		t.Fatal("ben sees anna's completions")
	}
	if err := repo.SetCompletion(ben, "h-admin", day, true); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben ticked off the admin's habit: %v", err)
	}
	if err := repo.SetCompletion(ben, "h-anna", day, false); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben unticked anna's invisible habit: %v", err)
	}
	if done, _ := repo.Completions(anna, day); !done["h-anna"][day] {
		t.Fatal("anna's completion is gone")
	}
	read.Name = "Mine"
	if err := repo.Save(ben, read); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben renamed the admin's habit: %v", err)
	}
	if err := repo.Delete(ben, "h-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's habit: %v", err)
	}
}

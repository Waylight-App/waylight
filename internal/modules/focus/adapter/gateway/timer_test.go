package gateway

import (
	"context"
	"testing"
	"time"

	"waylight/internal/modules/focus/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestTimerPerAccount: each account has its own timer and the stored state
// comes back as it was saved.
func TestTimerPerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "focus", Migrations()))
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)

	if _, ok, err := repo.Timer(anna); ok || err != nil {
		t.Fatalf("empty: %v %v", ok, err)
	}
	c, _ := domain.NewTimerConfig(50, 10, 20, 3)
	running := domain.NewTimer(c).Start(now)
	if err := repo.SaveTimer(anna, running); err != nil {
		t.Fatal(err)
	}
	got, ok, err := repo.Timer(anna)
	if err != nil || !ok || !got.Equal(running) {
		t.Fatalf("got %+v %v %v, want %+v", got, ok, err, running)
	}
	paused := running.Pause(now.Add(time.Minute))
	if err := repo.SaveTimer(anna, paused); err != nil {
		t.Fatal(err)
	}
	if got, _, _ = repo.Timer(anna); !got.Equal(paused) {
		t.Fatalf("update: %+v", got)
	}
	if _, ok, _ := repo.Timer(ben); ok {
		t.Fatal("ben sees anna's timer")
	}
}

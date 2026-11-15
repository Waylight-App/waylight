package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/onboarding/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestTourPerAccount: the owner finishing the tour does not hide it from a
// new member; guests follow the owner (ADR 0018).
func TestTourPerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "onboarding", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
	var done domain.State
	done.Complete(time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC))
	if err := repo.Save(admin, done); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.Load(anna); s.Completed() {
		t.Fatal("the owner's finished tour hides it from anna")
	}
	if s, _ := repo.Load(guest); !s.Completed() {
		t.Fatal("guests follow the owner")
	}
	if err := repo.Save(anna, done); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.Load(anna); !s.Completed() {
		t.Fatal("anna's own state was not kept")
	}
}

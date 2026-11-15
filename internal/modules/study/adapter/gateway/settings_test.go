package gateway

import (
	"context"
	"testing"

	"trisse-home/internal/modules/study/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestCreditsGoalPerAccount: the goal is personal, without falling back to
// the admin's (ADR 0018).
func TestCreditsGoalPerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "study", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if err := repo.SetTotalECTS(admin, 120); err != nil {
		t.Fatal(err)
	}
	if v, _ := repo.TotalECTS(anna); v != domain.DefaultTotalECTS {
		t.Fatalf("anna got %v", v)
	}
	if err := repo.SetTotalECTS(anna, 90); err != nil {
		t.Fatal(err)
	}
	if v, _ := repo.TotalECTS(admin); v != 120 {
		t.Fatalf("anna changed the admin's goal: %v", v)
	}
}

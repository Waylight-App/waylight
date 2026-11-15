package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/drop/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestDropIsPersonal: drop items move between an account's own devices, so
// each account sees and deletes only its own (ADR 0018).
func TestDropIsPersonal(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "drop", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	for ctx, id := range map[context.Context]string{admin: "d-admin", anna: "d-anna"} {
		if err := repo.Add(ctx, domain.Item{ID: id, Kind: domain.KindText, Text: id, Size: 7, CreatedAt: now, ExpiresAt: now.Add(time.Hour)}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if list, _ := repo.List(anna); len(list) != 1 || list[0].ID != "d-anna" {
		t.Fatalf("anna sees %v", list)
	}
	_ = repo.Delete(anna, "d-admin")
	if _, err := repo.Get(admin, "d-admin"); err != nil {
		t.Fatal("anna deleted the admin's item")
	}
	if n, _ := repo.DeleteExpired(anna, now.Add(2*time.Hour)); n != 1 {
		t.Fatalf("anna's cleanup removed %d items", n)
	}
	if n, _ := repo.DeleteExpired(context.Background(), now.Add(2*time.Hour)); n != 1 {
		t.Fatalf("the job removed %d items", n)
	}
}

package gateway

import (
	"context"
	"testing"

	"waylight/internal/modules/shopping/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestShoppingListPerAccount: members use the household (admin's) list until
// they connect their own, which never changes the admin's (ADR 0018).
func TestShoppingListPerAccount(t *testing.T) {
	repo := NewSettingsRepository(dbtest.Open(t, "shopping", Migrations()), prefixCipher{})
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if err := repo.Save(admin, domain.Settings{ListURL: "https://dav.example/household", Username: "home", Password: "x"}); err != nil {
		t.Fatal(err)
	}
	if s, err := repo.Load(anna); err != nil || s.ListURL != "https://dav.example/household" {
		t.Fatalf("anna uses the household list: %+v %v", s, err)
	}
	if err := repo.Save(anna, domain.Settings{ListURL: "https://dav.example/anna", Username: "anna", Password: "y"}); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.Load(admin); s.ListURL != "https://dav.example/household" {
		t.Fatalf("anna changed the admin's list: %+v", s)
	}
	if s, _ := repo.Load(anna); s.ListURL != "https://dav.example/anna" || s.Password != "y" {
		t.Fatalf("anna's own list: %+v", s)
	}
}

package gateway

import (
	"context"
	"testing"

	"waylight/internal/modules/todos/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestSourcesPerAccount: connected to-do lists belong to one account;
// members never see the admin's, guests do (ADR 0018).
func TestSourcesPerAccount(t *testing.T) {
	repo := NewSourceRepository(dbtest.Open(t, "todos", Migrations()), prefixCipher{})
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
	src := domain.Source{ID: "s-admin", Name: "Work", Kind: domain.KindCalDAV, URL: "https://dav.example/tasks/", Username: "admin", Password: "pw"}
	if err := repo.Save(admin, src); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(anna); len(list) != 0 {
		t.Fatalf("anna sees the admin's lists: %v", list)
	}
	if list, _ := repo.List(guest); len(list) != 1 {
		t.Fatalf("the guest sees %v", list)
	}
	if _, err := repo.Get(anna, "s-admin"); err == nil {
		t.Fatal("anna opened the admin's list")
	}
	src.URL = "https://evil.example/"
	if err := repo.Save(anna, src); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(admin, "s-admin"); got.URL != "https://dav.example/tasks/" {
		t.Fatalf("anna changed the admin's list to %q", got.URL)
	}
}

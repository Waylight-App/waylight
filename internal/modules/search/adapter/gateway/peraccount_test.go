package gateway

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/modules/search/usecase"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestSearchSettingsPerAccount: a member starts with the admin's shortcuts,
// the first change copies them into the member's own list, and from then on
// neither side affects the other (ADR 0018).
func TestSearchSettingsPerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "search", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
	sc := func(k string) domain.Shortcut {
		s, err := domain.NewShortcut(k, k, "https://example.org/?q={query}")
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	keywords := func(ctx context.Context) []string {
		list, err := repo.Shortcuts(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, s := range list {
			out = append(out, s.Keyword())
		}
		return out
	}
	if err := repo.AddShortcut(admin, sc("wiki")); err != nil {
		t.Fatal(err)
	}
	if got := keywords(anna); len(got) != 1 || got[0] != "wiki" {
		t.Fatalf("anna starts with the admin's shortcuts: %v", got)
	}
	// Anna's first change copies the list, then adds hers.
	if err := repo.AddShortcut(anna, sc("maps")); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteShortcut(anna, "wiki"); err != nil {
		t.Fatal(err)
	}
	if got := keywords(anna); len(got) != 1 || got[0] != "maps" {
		t.Fatalf("anna's own list: %v", got)
	}
	if got := keywords(admin); len(got) != 1 || got[0] != "wiki" {
		t.Fatalf("anna changed the admin's list: %v", got)
	}
	if got := keywords(guest); len(got) != 1 || got[0] != "wiki" {
		t.Fatalf("guests use the admin's list: %v", got)
	}
	// Deleting all own shortcuts does not bring the admin's back.
	if err := repo.DeleteShortcut(anna, "maps"); err != nil {
		t.Fatal(err)
	}
	if got := keywords(anna); len(got) != 0 {
		t.Fatalf("anna's emptied list shows %v", got)
	}
	if err := repo.DeleteShortcut(anna, "wiki"); !errors.Is(err, usecase.ErrShortcutNotFound) {
		t.Fatalf("anna deleted from the admin's list: %v", err)
	}
	// Config as code (the system) replaces the admin's shortcuts only.
	if err := repo.ReplaceShortcuts(context.Background(), []domain.Shortcut{sc("docs")}); err != nil {
		t.Fatal(err)
	}
	if got := keywords(admin); len(got) != 1 || got[0] != "docs" {
		t.Fatalf("admin after replace: %v", got)
	}
	if got := keywords(anna); len(got) != 0 {
		t.Fatalf("the replace touched anna: %v", got)
	}
	// Default provider follows the same rule.
	if err := repo.SetDefaultProvider(anna, "ddg"); err != nil {
		t.Fatal(err)
	}
	if p, _ := repo.DefaultProvider(admin); p == "ddg" {
		t.Fatal("anna changed the admin's default provider")
	}
}

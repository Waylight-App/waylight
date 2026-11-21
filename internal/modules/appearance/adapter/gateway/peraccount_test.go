package gateway

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"waylight/internal/modules/appearance/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestAppearancePerAccount: each account keeps its own appearance, starts
// from the admin's until it saves one, and guests see the admin's (ADR 0018).
func TestAppearancePerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "appearance", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})

	if got, err := repo.Load(anna); err != nil || got != domain.Default() {
		t.Fatalf("without any row: %+v %v", got, err)
	}
	german := domain.Default()
	if err := german.SetLanguage("de"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(admin, german); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Load(anna); got.Language() != "de" {
		t.Fatalf("anna starts from the admin's appearance, got %q", got.Language())
	}
	english := domain.Default()
	_ = english.SetLanguage("en")
	if err := repo.Save(anna, english); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		ctx  context.Context
		want string
	}{"admin": {admin, "de"}, "anna": {anna, "en"}, "guest": {guest, "de"}, "system": {context.Background(), "de"}} {
		if got, _ := repo.Load(c.ctx); got.Language() != c.want {
			t.Errorf("%s: language %q", name, got.Language())
		}
	}
	// A guest who could write (it cannot, the guard refuses) would only
	// ever touch the admin's row, never anna's.
	if got, _ := repo.Load(anna); got.Language() != "en" {
		t.Fatal("anna's appearance changed")
	}
}

// TestShowAdvancedIsNotInherited: a member who starts from the admin's
// appearance does not get the admin's advanced settings (#602).
func TestShowAdvancedIsNotInherited(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "appearance", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	a := domain.Default()
	a.SetShowAdvanced(true)
	if err := repo.Save(admin, a); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Load(admin); !got.ShowAdvanced() {
		t.Fatal("admin: not stored")
	}
	if got, _ := repo.Load(anna); got.ShowAdvanced() {
		t.Fatal("anna inherited the admin's advanced settings")
	}
	if err := repo.Save(anna, a); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Load(anna); !got.ShowAdvanced() {
		t.Fatal("anna: not stored")
	}
}

// Migration 0008 keeps the single row as the owner's.
func TestAppearanceMigrationKeepsTheRow(t *testing.T) {
	ctx := context.Background()
	before := fstest.MapFS{}
	entries, _ := fs.ReadDir(Migrations(), ".")
	for _, e := range entries {
		if e.Name() >= "0008" {
			continue
		}
		data, _ := fs.ReadFile(Migrations(), e.Name())
		before[e.Name()] = &fstest.MapFile{Data: data}
	}
	db := dbtest.Open(t, "appearance", before)
	if _, err := db.ExecContext(ctx, `INSERT INTO appearance_settings (id, theme, language, greeting_name, launcher, profile) VALUES (1, 'dark', 'de', 'Alex', 'both', 'everyday')`); err != nil {
		t.Fatal(err)
	}
	// 0008 and every later migration, as an update would run them.
	for _, e := range entries {
		if e.Name() < "0008" {
			continue
		}
		m, _ := fs.ReadFile(Migrations(), e.Name())
		if _, err := db.ExecContext(ctx, string(m)); err != nil {
			t.Fatalf("%s: %v", e.Name(), err)
		}
	}
	got, err := NewRepository(db).Load(ctx)
	if err != nil || got.Language() != "de" || got.GreetingName() != "Alex" || got.Theme() != "dark" {
		t.Fatalf("%+v %v", got, err)
	}
}

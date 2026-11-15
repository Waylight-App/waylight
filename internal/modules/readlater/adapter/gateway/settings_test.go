package gateway

import (
	"context"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestArchiveSettingPerAccount: each account keeps its own archive setting,
// falling back to the admin's, and the cleanup learns which accounts have
// articles (ADR 0018).
func TestArchiveSettingPerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "readlater", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if err := repo.SaveSettings(admin, domain.Settings{ArchiveDays: 30}); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.Settings(anna); s.ArchiveDays != 30 {
		t.Fatalf("anna falls back to the admin's setting, got %d", s.ArchiveDays)
	}
	if err := repo.SaveSettings(anna, domain.Settings{ArchiveDays: 7}); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.Settings(admin); s.ArchiveDays != 30 {
		t.Fatalf("anna changed the admin's setting: %d", s.ArchiveDays)
	}
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	for ctx, id := range map[context.Context]string{admin: "a1", anna: "a2"} {
		if err := repo.Save(ctx, domain.Article{ID: id, URL: "https://example.org/" + id, Status: domain.StatusUnread, SavedAt: now}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if owners, _ := repo.Owners(context.Background()); !slices.Equal(owners, []string{"anna", "owner"}) {
		t.Fatalf("owners %v", owners)
	}
	if owners, _ := repo.Owners(anna); !slices.Equal(owners, []string{"anna"}) {
		t.Fatalf("an account only learns about itself: %v", owners)
	}
}

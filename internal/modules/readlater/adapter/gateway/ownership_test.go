package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/readlater/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestReadingListIsPersonal: each account reads its own list and only the
// owner changes an article's status (ADR 0018).
func TestReadingListIsPersonal(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "readlater", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
	mk := func(ctx context.Context, id string) {
		t.Helper()
		if err := repo.Save(ctx, domain.Article{ID: id, URL: "https://example.org/" + id, Title: id, Status: domain.StatusUnread, SavedAt: now}, nil); err != nil {
			t.Fatal(err)
		}
	}
	mk(admin, "a-admin")
	mk(anna, "a-anna")
	for name, c := range map[string]struct {
		ctx  context.Context
		want string
	}{"admin": {admin, "a-admin"}, "anna": {anna, "a-anna"}, "guest": {guest, "a-admin"}} {
		if list, err := repo.List(c.ctx); err != nil || len(list) != 1 || list[0].ID != c.want {
			t.Errorf("%s: %v %v", name, list, err)
		}
	}
	if err := repo.SetStatus(anna, "a-admin", domain.StatusRead, now); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("anna marked the admin's article: %v", err)
	}
	if err := repo.Save(anna, domain.Article{ID: "a-admin", URL: "x", Title: "Mine", Status: domain.StatusUnread, SavedAt: now}, nil); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("anna overwrote the admin's article: %v", err)
	}
	_ = repo.Delete(anna, "a-admin")
	if _, err := repo.Get(admin, "a-admin"); err != nil {
		t.Fatal("anna deleted the admin's article")
	}
	if n, _ := repo.TotalBytes(anna); n != 0 {
		t.Fatalf("total bytes %d", n) // both articles are empty; the sum covers the instance
	}
}

package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/rss/domain"
	"trisse-home/internal/modules/rss/usecase"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestSubscriptionsArePersonal: two accounts can subscribe to the same feed,
// each sees its own items and read state, and only the owner marks read or
// unsubscribes (ADR 0018).
func TestSubscriptionsArePersonal(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "rss", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	system := context.Background()
	const url = "https://news.example/feed.xml"

	if err := repo.CreateFeed(admin, domain.Feed{ID: "f-admin", URL: url, Title: "News", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateFeed(anna, domain.Feed{ID: "f-anna", URL: url, Title: "News", CreatedAt: now}); err != nil {
		t.Fatalf("anna cannot subscribe to the same feed: %v", err)
	}
	if err := repo.CreateFeed(anna, domain.Feed{ID: "f-anna2", URL: url, Title: "News", CreatedAt: now}); !errors.Is(err, domain.ErrAlreadySubscribed) {
		t.Fatalf("a second subscription for anna: %v", err)
	}
	// The fetch job (system) fills both feeds.
	for _, f := range []string{"f-admin", "f-anna"} {
		if _, err := repo.AddItems(system, []domain.Item{{ID: f + "-1", FeedID: f, GUID: "g1", Title: "Story", PublishedAt: now}}); err != nil {
			t.Fatal(err)
		}
	}
	items := func(ctx context.Context) []domain.Item {
		list, err := repo.ListItems(ctx, usecase.ItemFilter{Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		return list
	}
	if got := items(anna); len(got) != 1 || got[0].FeedID != "f-anna" {
		t.Fatalf("anna sees %v", got)
	}
	// Reading in her feed does not mark the admin's copy.
	if err := repo.MarkAllRead(anna, ""); err != nil {
		t.Fatal(err)
	}
	if got := items(admin); got[0].Read {
		t.Fatal("anna's mark-all-read changed the admin's read state")
	}
	if err := repo.SetRead(anna, "f-admin-1", true); !errors.Is(err, domain.ErrItemNotFound) {
		t.Fatalf("anna marked an item of the admin's feed: %v", err)
	}
	if counts, _ := repo.UnreadCounts(admin); counts["f-admin"] != 1 || counts["f-anna"] != 0 {
		t.Fatalf("admin's unread counts %v", counts)
	}
	_ = repo.DeleteFeed(anna, "f-admin")
	if _, err := repo.GetFeed(admin, "f-admin"); err != nil {
		t.Fatal("anna unsubscribed the admin")
	}
	// Anna cannot push items into the admin's feed.
	if n, _ := repo.AddItems(anna, []domain.Item{{ID: "x", FeedID: "f-admin", GUID: "g2", Title: "Spam", PublishedAt: now}}); n != 0 {
		t.Fatal("anna added an item to the admin's feed")
	}
}

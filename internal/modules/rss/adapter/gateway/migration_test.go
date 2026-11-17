package gateway

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	"waylight/internal/platform/database/dbtest"
)

// Migration 0003 rebuilds rss_feeds with foreign keys on. Items and their
// read state must survive, and the cascade must still work afterwards.
func TestOwnerMigrationKeepsItems(t *testing.T) {
	ctx := context.Background()
	before := fstest.MapFS{}
	for _, n := range []string{"0001_init.sql", "0002_category.sql"} {
		data, err := fs.ReadFile(Migrations(), n)
		if err != nil {
			t.Fatal(err)
		}
		before[n] = &fstest.MapFile{Data: data}
	}
	db := dbtest.Open(t, "rss", before)
	if _, err := db.ExecContext(ctx, `INSERT INTO rss_feeds (id, url, title, created_at, category) VALUES ('f1', 'https://a.example/feed', 'A', '2026-01-01T00:00:00Z', 'news');
		INSERT INTO rss_items (id, feed_id, guid, title, published_at, read) VALUES ('i1', 'f1', 'g1', 'One', '2026-01-02T00:00:00Z', 1), ('i2', 'f1', 'g2', 'Two', '2026-01-03T00:00:00Z', 0)`); err != nil {
		t.Fatal(err)
	}
	m, err := fs.ReadFile(Migrations(), "0003_owner.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(m)); err != nil {
		t.Fatal(err)
	}
	var items, read int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*), SUM(read) FROM rss_items`).Scan(&items, &read); err != nil || items != 2 || read != 1 {
		t.Fatalf("items %d, read %d, %v", items, read, err)
	}
	var owner, category string
	if err := db.QueryRowContext(ctx, `SELECT owner_id, category FROM rss_feeds WHERE id = 'f1'`).Scan(&owner, &category); err != nil || owner != "owner" || category != "news" {
		t.Fatalf("feed: %q %q %v", owner, category, err)
	}
	// Two accounts may subscribe to the same URL now.
	if _, err := db.ExecContext(ctx, `INSERT INTO rss_feeds (id, url, title, created_at, owner_id) VALUES ('f2', 'https://a.example/feed', 'A', '2026-01-01T00:00:00Z', 'anna')`); err != nil {
		t.Fatal(err)
	}
	// The cascade points at the renamed table.
	if _, err := db.ExecContext(ctx, `DELETE FROM rss_feeds WHERE id = 'f1'`); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM rss_items`).Scan(&items); err != nil || items != 0 {
		t.Fatalf("items after deleting the feed: %d %v", items, err)
	}
}

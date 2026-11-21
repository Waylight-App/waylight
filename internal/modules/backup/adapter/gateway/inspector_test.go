package gateway

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"waylight/internal/modules/backup/domain"
	"waylight/internal/platform/database/dbtest"
)

func TestInspector(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "backup", Migrations())
	_, _ = db.Exec(`INSERT INTO schema_migrations (module, version, applied_at) VALUES ('auth', '0001', '2026-01-01T00:00:00Z')`)
	insp := NewInspector(db)
	dir := t.TempDir()

	snapshot := filepath.Join(dir, "snap.db")
	if err := NewSnapshotter(db).SnapshotTo(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	info, err := insp.Inspect(ctx, snapshot)
	if err != nil || info.Integrity != "ok" || len(info.Migrations) != 2 {
		t.Fatalf("snapshot: %+v %v", info, err)
	}
	current, _ := insp.Current(ctx)
	if err := domain.ValidateRestore(info, current); err != nil {
		t.Fatalf("own snapshot must be restorable: %v", err)
	}

	garbage := filepath.Join(dir, "garbage.db")
	_ = os.WriteFile(garbage, []byte("this is not sqlite at all, just some text that is long enough"), 0o600)
	if _, err := insp.Inspect(ctx, garbage); err == nil {
		t.Fatal("garbage accepted")
	}

	// An empty SQLite database has no schema_migrations table.
	empty := filepath.Join(dir, "empty.db")
	_, _ = db.Exec(`VACUUM INTO ?`, filepath.Join(dir, "tmp.db"))
	other := dbtest.Open(t, "x", nil)
	_, _ = other.Exec(`DROP TABLE schema_migrations`)
	_, _ = other.Exec(`VACUUM INTO ?`, empty)
	info, err = insp.Inspect(ctx, empty)
	if err != nil || len(info.Migrations) != 0 || domain.ValidateRestore(info, current) == nil {
		t.Fatalf("foreign db: %+v %v", info, err)
	}
	// The live database must still work and have nothing attached.
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_database_list WHERE name = 'restore_check'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("still attached: %d %v", n, err)
	}
}

// A stored backup is inspected in place, so the inspector must not touch
// it: same bytes afterwards and no journal files next to it, even for a
// copy of a database in WAL mode and a path that needs escaping.
func TestInspectorLeavesTheFileUntouched(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "backup", Migrations())
	for _, q := range []string{
		`PRAGMA journal_mode = WAL`,
		`INSERT INTO schema_migrations (module, version, applied_at) VALUES ('auth', '0001', '2026-01-01T00:00:00Z')`,
		`CREATE TABLE dashboard_tabs (id TEXT PRIMARY KEY)`,
		`CREATE TABLE dashboard_widgets (id TEXT PRIMARY KEY)`,
		`INSERT INTO dashboard_tabs VALUES ('home'), ('work')`,
		`INSERT INTO dashboard_widgets VALUES ('a'), ('b'), ('c')`,
	} {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	dir := filepath.Join(t.TempDir(), "my backups #1")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "waylight-20261121-030000.db")
	if err := NewSnapshotter(db).SnapshotTo(ctx, path); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(path)

	info, err := NewInspector(db).Inspect(ctx, path)
	if err != nil || info.Integrity != "ok" || info.Tabs != 2 || info.Widgets != 3 {
		t.Fatalf("inspect: %+v %v", info, err)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("the backup file changed")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("files next to the backup: %v", entries)
	}

	// The attach itself refuses writes.
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `ATTACH DATABASE ? AS ro_check`, readOnlyURI(path)); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE TABLE ro_check.x (y)`); err == nil {
		t.Fatal("a write to the attached backup succeeded")
	}
	_, _ = conn.ExecContext(ctx, `DETACH DATABASE ro_check`)
}

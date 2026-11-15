package gateway

import (
	"context"
	"errors"
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/platform/database/dbtest"
)

// Migration 0010 rebuilds auth_accounts for the member role: owner, guests,
// kiosks and their state must survive.
func TestMembersMigrationKeepsAccounts(t *testing.T) {
	ctx := context.Background()
	before := fstest.MapFS{}
	entries, err := fs.ReadDir(Migrations(), ".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() >= "0010" {
			continue
		}
		data, err := fs.ReadFile(Migrations(), e.Name())
		if err != nil {
			t.Fatal(err)
		}
		before[e.Name()] = &fstest.MapFile{Data: data}
	}
	db := dbtest.Open(t, "auth", before)
	if _, err := db.ExecContext(ctx, `INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at, totp_secret_enc, disabled, tabs)
		VALUES ('owner', 'owner', 'owner', 'h1', '2026-11-01T00:00:00Z', '2026-11-01T00:00:00Z', 'secret', 0, '[]'),
		       ('g1', 'guest', 'anna', 'h2', '2026-11-02T00:00:00Z', '2026-11-02T00:00:00Z', '', 1, '["t1"]'),
		       ('k1', 'kiosk', 'flur', '!kiosk', '2026-11-03T00:00:00Z', '2026-11-03T00:00:00Z', '', 0, '["t2"]')`); err != nil {
		t.Fatal(err)
	}
	m, err := fs.ReadFile(Migrations(), "0010_members_groups.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(m)); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM auth_accounts WHERE (id = 'owner' AND totp_secret_enc = 'secret') OR (id = 'g1' AND disabled = 1) OR (id = 'k1' AND tabs = '["t2"]')`).Scan(&n); err != nil || n != 3 {
		t.Fatalf("accounts after migration: %d %v", n, err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at) VALUES ('m1', 'member', 'ben', 'h', '', '')`); err != nil {
		t.Fatal("member role allowed", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO auth_accounts (id, role, name, password_hash, created_at, updated_at) VALUES ('o2', 'owner', 'x', 'h', '', '')`); err == nil {
		t.Fatal("still only one owner")
	}
}

func TestGroupRepository(t *testing.T) {
	ctx := context.Background()
	r := NewGroupRepository(dbtest.Open(t, "auth", Migrations()))
	now := time.Date(2026, 11, 15, 9, 0, 0, 0, time.UTC)
	office, _ := domain.NewGroup("g1", "Office", []string{"m1", "m2"}, now)
	home, _ := domain.NewGroup("g2", "home", []string{"m1"}, now)
	for _, g := range []domain.Group{office, home} {
		if err := r.Create(ctx, g); err != nil {
			t.Fatal(err)
		}
	}
	dup, _ := domain.NewGroup("g3", "OFFICE", nil, now)
	if err := r.Create(ctx, dup); !errors.Is(err, domain.ErrGroupNameTaken) {
		t.Fatalf("names are unique regardless of case: %v", err)
	}
	list, err := r.List(ctx)
	if err != nil || len(list) != 2 || list[0].Name() != "home" || !slices.Equal(list[1].Accounts(), []string{"m1", "m2"}) {
		t.Fatalf("%+v %v", list, err)
	}
	if ids, _ := r.GroupsOf(ctx, "m1"); !slices.Equal(ids, []string{"g1", "g2"}) {
		t.Fatalf("groups of m1: %v", ids)
	}
	_ = office.Rename("Büro")
	office.SetAccounts([]string{"m2"})
	if err := r.Update(ctx, office); err != nil {
		t.Fatal(err)
	}
	if got, err := r.Get(ctx, "g1"); err != nil || got.Name() != "Büro" || !slices.Equal(got.Accounts(), []string{"m2"}) || !got.CreatedAt().Equal(now) {
		t.Fatalf("%+v %v", got, err)
	}
	if err := r.RemoveAccount(ctx, "m1"); err != nil {
		t.Fatal(err)
	}
	if ids, _ := r.GroupsOf(ctx, "m1"); len(ids) != 0 {
		t.Fatalf("m1 still in %v", ids)
	}
	if err := r.Delete(ctx, "g1"); err != nil {
		t.Fatal(err)
	}
	if ids, _ := r.GroupsOf(ctx, "m2"); len(ids) != 0 {
		t.Fatalf("deleting a group removes its accounts: %v", ids)
	}
	if err := r.Delete(ctx, "g1"); !errors.Is(err, domain.ErrGroupNotFound) {
		t.Fatal(err)
	}
	ghost, _ := domain.NewGroup("nope", "Ghost", nil, now)
	if err := r.Update(ctx, ghost); !errors.Is(err, domain.ErrGroupNotFound) {
		t.Fatal(err)
	}
}

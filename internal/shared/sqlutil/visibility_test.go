package sqlutil

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	_ "modernc.org/sqlite"

	"trisse-home/internal/shared/kernel"
)

// TestVisibleToMatchesTheKernelRule runs the SQL filter against a table that
// holds every combination and compares the result with kernel.VisibleTo, so
// the two forms of the rule cannot drift apart.
func TestVisibleToMatchesTheKernelRule(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `CREATE TABLE notes (id TEXT, owner_id TEXT, visibility TEXT)`); err != nil {
		t.Fatal(err)
	}
	type rec struct {
		id, owner string
		v         kernel.Visibility
	}
	var recs []rec
	for _, owner := range []string{"anna", "ben", "owner"} {
		for _, v := range []kernel.Visibility{kernel.Private, kernel.Everyone, kernel.GroupVisibility("office"), kernel.GroupVisibility("home")} {
			r := rec{owner + "/" + string(v), owner, v}
			recs = append(recs, r)
			if _, err := db.ExecContext(ctx, `INSERT INTO notes VALUES (?, ?, ?)`, r.id, r.owner, string(r.v)); err != nil {
				t.Fatal(err)
			}
		}
	}
	principals := []kernel.Principal{
		{AccountID: "anna", Role: "member", Groups: []string{"office"}},
		{AccountID: "ben", Role: "member", Groups: []string{"office", "home"}},
		{AccountID: "owner", Role: "owner"},
		{AccountID: "g1", Role: "guest"},
		kernel.PrincipalFrom(ctx),
	}
	for _, p := range principals {
		got := visibleIDs(t, db, p)
		var want []string
		for _, r := range recs {
			if r.v.VisibleTo(p, r.owner) {
				want = append(want, r.id)
			}
		}
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Errorf("%s/%s:\n got  %v\n want %v", p.Role, p.AccountID, got, want)
		}
	}
}

func TestOwnedBy(t *testing.T) {
	if cond, args := OwnedBy(kernel.Principal{AccountID: "anna", Role: "member"}, ""); cond != "owner_id = ?" || args[0] != "anna" {
		t.Fatal(cond, args)
	}
	if cond, args := OwnedBy(kernel.PrincipalFrom(context.Background()), "t"); cond != "1 = 1" || args != nil {
		t.Fatal(cond, args)
	}
}

func visibleIDs(t *testing.T, db *sql.DB, p kernel.Principal) []string {
	t.Helper()
	cond, args := VisibleTo(p, "n")
	rows, err := db.QueryContext(context.Background(), `SELECT n.id FROM notes n WHERE `+cond+` ORDER BY n.id`, args...)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return ids
}

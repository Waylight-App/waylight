package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/notes/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestNoteOwnership proves ADR 0018 for notes: a member's notes are private
// by default, nobody else reads, finds, changes or deletes them, and notes
// shared with everyone are readable but stay their owner's to change.
func TestNoteOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "notes", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})

	secret, _ := domain.NewNote("n-anna", "Diary", "private words", false, now)
	if err := repo.Save(anna, secret); err != nil {
		t.Fatal(err)
	}
	shopping, _ := domain.NewNote("n-admin", "Shopping", "milk", false, now)
	if err := repo.Save(admin, shopping); err != nil {
		t.Fatal(err)
	}

	// Reading: the admin's note is for everyone (today's default), Anna's
	// is private, also towards the admin and guests.
	for name, c := range map[string]struct {
		ctx  context.Context
		want int
	}{"anna": {anna, 2}, "ben": {ben, 1}, "admin": {admin, 1}, "guest": {guest, 1}, "system": {context.Background(), 2}} {
		list, err := repo.List(c.ctx)
		if err != nil || len(list) != c.want {
			t.Errorf("%s lists %d notes: %v", name, len(list), err)
		}
	}
	for _, ctx := range []context.Context{ben, admin, guest} {
		if _, err := repo.Get(ctx, "n-anna"); !errors.Is(err, domain.ErrNoteNotFound) {
			t.Fatalf("read anna's note: %v", err)
		}
		if hits, _ := repo.Search(ctx, "private", 10); len(hits) != 0 {
			t.Fatalf("found anna's note: %v", hits)
		}
	}
	if hits, _ := repo.Search(anna, "private", 10); len(hits) != 1 {
		t.Fatalf("anna cannot find her own note: %v", hits)
	}

	// Changing: only the owner, whatever the visibility.
	stolen := domain.RestoreNote("n-anna", "Mine", "x", false, now, now).WithVersion(2)
	if err := repo.Save(ben, stolen); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's note: %v", err)
	}
	edited, _ := repo.Get(ben, "n-admin")
	_ = edited.Edit("Shopping", "milk, beer", false, now)
	edited.Bump()
	if err := repo.Save(ben, edited); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed the admin's shared note: %v", err)
	}
	if err := repo.AddRevision(ben, domain.Revision{ID: "r1", NoteID: "n-admin", Title: "x", Body: "y", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if revs, _ := repo.ListRevisions(admin, "n-admin"); len(revs) != 0 {
		t.Fatalf("a refused edit left a revision: %v", revs)
	}
	if err := repo.Delete(ben, "n-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's note: %v", err)
	}
	if err := repo.Delete(ben, "n-anna"); err != nil {
		t.Fatalf("deleting an invisible note must look like a missing one: %v", err)
	}
	if got, err := repo.Get(anna, "n-anna"); err != nil || got.Title() != "Diary" {
		t.Fatalf("anna's note after ben's attempts: %+v %v", got, err)
	}

	// The owner keeps full control, including a version conflict.
	mine, _ := repo.Get(anna, "n-anna")
	_ = mine.Edit("Diary", "new words", false, now)
	mine.Bump()
	if err := repo.Save(anna, mine); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(anna, mine); !errors.Is(err, domain.ErrVersionConflict) {
		t.Fatalf("stale save: %v", err)
	}
	if err := repo.AddRevision(anna, domain.Revision{ID: "r2", NoteID: "n-anna", Title: "Diary", Body: "private words", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if revs, _ := repo.ListRevisions(ben, "n-anna"); len(revs) != 0 {
		t.Fatal("ben read anna's history")
	}
	if revs, _ := repo.ListRevisions(anna, "n-anna"); len(revs) != 1 {
		t.Fatalf("anna's history: %v", revs)
	}
	if err := repo.Delete(anna, "n-anna"); err != nil {
		t.Fatal(err)
	}
}

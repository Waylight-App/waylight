package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/board/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

var now = time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)

func note(t *testing.T, id, expires string, v kernel.Visibility) domain.Note {
	t.Helper()
	n, err := domain.New(id, domain.Input{Text: "Note " + id, ExpiresOn: expires}, kernel.Access{Visibility: v}, now, now)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// TestOwnership proves ADR 0018 for the board.
func TestOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "board", Migrations()))
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	shared, private := note(t, "shared", "2026-11-30", kernel.Everyone), note(t, "private", "", kernel.Private)
	for _, n := range []domain.Note{shared, private} {
		if err := repo.Save(anna, n); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := repo.List(ben)
	if len(list) != 1 || list[0].ID != "shared" || list[0].ExpiryDate() != "2026-11-30" || list[0].Access.Owner != "anna" {
		t.Fatalf("ben sees %+v", list)
	}
	if _, err := repo.Get(ben, "private"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben read anna's private note: %v", err)
	}
	shared.Pinned = true
	if err := repo.Save(ben, shared); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben pinned anna's note: %v", err)
	}
	if err := repo.Delete(ben, "shared"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted anna's note: %v", err)
	}
}

func TestDeleteExpired(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "board", Migrations()))
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	for ctx, n := range map[context.Context]domain.Note{anna: note(t, "a", "2026-11-23", kernel.Private), ben: note(t, "b", "2026-11-24", kernel.Private)} {
		if err := repo.Save(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.Save(anna, note(t, "keep", "", kernel.Private)); err != nil {
		t.Fatal(err)
	}
	n, err := repo.DeleteExpired(context.Background(), time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC))
	if err != nil || n != 2 {
		t.Fatalf("%d %v", n, err)
	}
	if list, _ := repo.List(anna); len(list) != 1 || list[0].ID != "keep" {
		t.Fatalf("%+v", list)
	}
}

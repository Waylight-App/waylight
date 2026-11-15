package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestTodoOwnership proves ADR 0018 for to-dos: a member's to-dos are
// private by default and only their owner changes or deletes a to-do.
func TestTodoOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "todos", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})

	mine, _ := domain.NewTodo("t-anna", "Call the doctor", time.Time{}, 0, now)
	if err := repo.Save(anna, mine); err != nil {
		t.Fatal(err)
	}
	family, _ := domain.NewTodo("t-admin", "Buy milk", time.Time{}, 1, now)
	if err := repo.Save(admin, family); err != nil {
		t.Fatal(err)
	}
	done, _ := domain.NewTodo("t-admin-done", "Taxes", time.Time{}, 2, now)
	if err := repo.Save(admin, done); err != nil {
		t.Fatal(err)
	}

	for name, c := range map[string]struct {
		ctx  context.Context
		want int
	}{"anna": {anna, 3}, "ben": {ben, 2}, "admin": {admin, 2}, "guest": {guest, 2}, "system": {context.Background(), 3}} {
		if list, err := repo.List(c.ctx); err != nil || len(list) != c.want {
			t.Errorf("%s lists %d to-dos: %v", name, len(list), err)
		}
	}
	if _, err := repo.Get(ben, "t-anna"); !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatalf("ben read anna's to-do: %v", err)
	}

	shared, _ := repo.Get(ben, "t-admin")
	_ = shared.Edit("Buy beer", time.Time{})
	shared.Bump()
	if err := repo.Save(ben, shared); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed the admin's to-do: %v", err)
	}
	if err := repo.Delete(ben, "t-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's to-do: %v", err)
	}
	if err := repo.Delete(ben, "t-anna"); err != nil {
		t.Fatalf("an invisible to-do must look missing: %v", err)
	}
	// Clearing several (as "clear completed" does) skips foreign to-dos.
	if err := repo.Delete(ben, "t-admin", "t-admin-done"); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(admin); len(list) != 2 {
		t.Fatalf("ben's clear removed the admin's to-dos: %d left", len(list))
	}

	// Background jobs (reminders) act as the system and may update any.
	sys, _ := repo.Get(context.Background(), "t-anna")
	sys.Bump()
	if err := repo.Save(context.Background(), sys); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.Get(anna, "t-anna"); err != nil || got.Version() != sys.Version() {
		t.Fatalf("the system update changed the owner: %+v %v", got, err)
	}
}

package gateway

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/focus/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestFocusIsPersonal: focus statistics count only the account's own
// sessions, even when the admin's are visible to everyone (ADR 0018);
// guests see the shared ones.
func TestFocusIsPersonal(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "focus", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})

	for ctx, id := range map[context.Context]string{admin: "s-admin", anna: "s-anna"} {
		s, _ := domain.NewSession(id, 25, now, now)
		if err := repo.Save(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	since := now.Add(-time.Hour)
	for name, c := range map[string]struct {
		ctx  context.Context
		want string
	}{"admin": {admin, "s-admin"}, "anna": {anna, "s-anna"}, "guest": {guest, "s-admin"}} {
		got, err := repo.Since(c.ctx, since)
		if err != nil || len(got) != 1 || got[0].ID != c.want {
			t.Errorf("%s: %v %v", name, got, err)
		}
	}
	// Anna pruning her history leaves the admin's alone; the job prunes all.
	if err := repo.DeleteBefore(anna, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Since(admin, since); len(got) != 1 {
		t.Fatal("anna pruned the admin's sessions")
	}
	if err := repo.DeleteBefore(context.Background(), now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Since(context.Background(), since); len(got) != 0 {
		t.Fatalf("the job left %v", got)
	}
}

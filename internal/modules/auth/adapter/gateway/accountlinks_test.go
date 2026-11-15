package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestAccountLinkRepository(t *testing.T) {
	ctx := context.Background()
	r := NewAccountLinkRepository(dbtest.Open(t, "auth", Migrations()))
	now := time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC)
	inv, err := domain.NewInvite("i1", domain.HashToken("invite"), "lena", domain.RoleGuest, []string{"home"}, 7, now)
	if err != nil {
		t.Fatal(err)
	}
	reset := domain.AccountLink{ID: "r1", TokenHash: domain.HashToken("reset"), Purpose: domain.LinkReset, AccountID: "g1", Name: "oma", Role: domain.RoleGuest, CreatedAt: now, ExpiresAt: now.Add(time.Hour)}
	for _, l := range []domain.AccountLink{inv, reset} {
		if err := r.Save(ctx, l); err != nil {
			t.Fatal(err)
		}
	}
	got, err := r.FindByTokenHash(ctx, domain.HashToken("invite"))
	if err != nil || got.Name != "lena" || got.Role != domain.RoleGuest || len(got.Tabs) != 1 || !got.ExpiresAt.Equal(inv.ExpiresAt) {
		t.Fatalf("%+v %v", got, err)
	}
	if list, _ := r.List(ctx, domain.LinkInvite); len(list) != 1 || list[0].ID != "i1" {
		t.Fatalf("invites: %+v", list)
	}

	// A link is consumed exactly once.
	if ok, err := r.Delete(ctx, "i1"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	if ok, _ := r.Delete(ctx, "i1"); ok {
		t.Fatal("a second delete must report that the link was gone")
	}
	if _, err := r.FindByTokenHash(ctx, domain.HashToken("invite")); !errors.Is(err, domain.ErrInvalidAccountLink) {
		t.Fatal(err)
	}

	if err := r.DeleteExpired(ctx, now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := r.FindByTokenHash(ctx, domain.HashToken("reset")); !errors.Is(err, domain.ErrInvalidAccountLink) {
		t.Fatal("expired links are removed", err)
	}
}

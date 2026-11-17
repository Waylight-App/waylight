package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/platform/database/dbtest"
)

func TestIdentityRepository(t *testing.T) {
	ctx := context.Background()
	r := NewIdentityRepository(dbtest.Open(t, "auth", Migrations()))
	at := time.Date(2026, 11, 17, 8, 0, 0, 0, time.UTC)
	l := domain.IdentityLink{AccountID: domain.OwnerID, Issuer: "https://login.example", Subject: "s1", Email: "me@outlook.com", CreatedAt: at}
	if err := r.Add(ctx, l); err != nil {
		t.Fatal(err)
	}
	// One identity, one account.
	if err := r.Add(ctx, domain.IdentityLink{AccountID: "m1", Issuer: l.Issuer, Subject: "s1", CreatedAt: at}); !errors.Is(err, domain.ErrIdentityTaken) {
		t.Fatalf("taken: %v", err)
	}
	// The same subject at another issuer is another person.
	if err := r.Add(ctx, domain.IdentityLink{AccountID: "m1", Issuer: "https://other.example", Subject: "s1", CreatedAt: at}); err != nil {
		t.Fatal(err)
	}
	if got, err := r.Find(ctx, l.Issuer, "s1"); err != nil || got != domain.OwnerID {
		t.Fatalf("find: %q %v", got, err)
	}
	list, err := r.ListByAccount(ctx, domain.OwnerID)
	if err != nil || len(list) != 1 || list[0].Email != "me@outlook.com" || !list[0].CreatedAt.Equal(at) {
		t.Fatalf("list: %+v %v", list, err)
	}
	if err := r.Remove(ctx, "m1", l.Issuer, "s1"); !errors.Is(err, domain.ErrIdentityNotFound) {
		t.Fatalf("someone else's link removed: %v", err)
	}
	if err := r.RemoveAccount(ctx, "m1"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Find(ctx, "https://other.example", "s1"); !errors.Is(err, domain.ErrIdentityNotFound) {
		t.Fatalf("after account removal: %v", err)
	}
	if err := r.Remove(ctx, domain.OwnerID, l.Issuer, "s1"); err != nil {
		t.Fatal(err)
	}
}

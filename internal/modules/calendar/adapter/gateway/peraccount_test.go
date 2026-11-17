package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/calendar/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestCalendarSourcesPerAccount: subscriptions and CalDAV settings belong to
// one account, members get no fallback to the admin's, guests use the
// admin's (ADR 0018).
func TestCalendarSourcesPerAccount(t *testing.T) {
	db := dbtest.Open(t, "calendar", Migrations())
	subs := NewSubscriptionRepository(db, plainCipher{})
	settings := NewSettingsRepository(db, plainCipher{})
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})

	for ctx, id := range map[context.Context]string{admin: "s-admin", anna: "s-anna"} {
		if err := subs.Save(ctx, domain.Subscription{ID: id, Name: id, URL: "https://example.org/" + id + ".ics", Color: "#123456", RefreshInterval: time.Hour}); err != nil {
			t.Fatal(err)
		}
	}
	for name, c := range map[string]struct {
		ctx  context.Context
		want string
	}{"admin": {admin, "s-admin"}, "anna": {anna, "s-anna"}, "guest": {guest, "s-admin"}, "system": {context.Background(), "s-admin"}} {
		if list, err := subs.List(c.ctx); err != nil || len(list) != 1 || list[0].ID != c.want {
			t.Errorf("%s: %v %v", name, list, err)
		}
	}
	if _, err := subs.Get(anna, "s-admin"); !errors.Is(err, domain.ErrSubscriptionNotFound) {
		t.Fatalf("anna read the admin's subscription: %v", err)
	}
	if err := subs.Save(anna, domain.Subscription{ID: "s-admin", Name: "Taken", URL: "https://evil.example/x.ics", Color: "#000000", RefreshInterval: time.Hour}); err != nil {
		t.Fatal(err)
	}
	if got, _ := subs.Get(admin, "s-admin"); got.Name != "s-admin" {
		t.Fatalf("anna overwrote the admin's subscription: %+v", got)
	}
	_ = subs.Delete(anna, "s-admin")
	if _, err := subs.Get(admin, "s-admin"); err != nil {
		t.Fatal("anna deleted the admin's subscription")
	}

	s, err := domain.NewSettings([]string{"https://dav.example/cal/"}, "admin", "pw")
	if err != nil {
		t.Fatal(err)
	}
	if err := settings.Save(admin, s); err != nil {
		t.Fatal(err)
	}
	if _, err := settings.Load(anna); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("anna must not use the admin's CalDAV account: %v", err)
	}
	if got, err := settings.Load(guest); err != nil || got.Username != "admin" {
		t.Fatalf("guests use the admin's: %+v %v", got, err)
	}
}

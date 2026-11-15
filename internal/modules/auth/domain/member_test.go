package domain

import (
	"errors"
	"testing"
	"time"
)

func TestMemberMayAccess(t *testing.T) {
	kinds := map[string]string{"notes": KindPersonalContent, "appearance": KindPersonalSettings, "services": "instance", "weather": "instance"}
	cases := []struct {
		method, path string
		want         bool
	}{
		{"GET", "/api/auth/status", true},
		{"PUT", "/api/auth/password", true},
		{"DELETE", "/api/auth/sessions/s1", true},
		{"GET", "/api/auth/guests", false},
		{"POST", "/api/auth/members", false},
		{"GET", "/api/auth/tokens", false},
		{"POST", "/api/notes", true},
		{"DELETE", "/api/notes/n1", true},
		{"PUT", "/api/appearance", true},
		{"GET", "/api/services", true},
		{"POST", "/api/services/s1/wake", false},
		{"GET", "/api/weather/settings", false},
		{"PUT", "/api/weather/settings", false},
		{"GET", "/api/modules", true},
		{"GET", "/api/transfer/export", false},
		{"GET", "/api/backups", false},
		{"GET", "/api/notifications/channels", false},
		{"POST", "/api/unknownmodule/x", false},
		{"GET", "/healthz", false},
		{"GET", "/api/", false},
	}
	for _, c := range cases {
		if got := MemberMayAccess(c.method, c.path, kinds); got != c.want {
			t.Errorf("%s %s: got %v", c.method, c.path, got)
		}
	}
}

func TestNewMember(t *testing.T) {
	now := time.Now()
	m, err := NewMember("m1", "anna", "hash", now)
	if err != nil || m.Role() != RoleMember || m.IsAdmin() || !m.CanSee("any-tab") {
		t.Fatalf("%+v %v", m, err)
	}
	for _, bad := range [][3]string{{"m1", "owner", "h"}, {"owner", "anna", "h"}, {"m1", "Anna!", "h"}, {"m1", "anna", ""}} {
		if _, err := NewMember(bad[0], bad[1], bad[2], now); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
	if err := m.SetMemberAccess(true, now); err != nil || !m.Disabled() {
		t.Fatal(err)
	}
	g, _ := NewGuest("g1", "gast", "h", nil, now)
	if err := g.SetMemberAccess(true, now); !errors.Is(err, ErrMemberNotFound) {
		t.Fatal("a guest is no member", err)
	}
}

func TestGroup(t *testing.T) {
	g, err := NewGroup("g1", "  Office ", []string{"b", "a", "owner", "", "a"}, time.Now())
	if err != nil || g.Name() != "Office" {
		t.Fatal(g, err)
	}
	if got := g.Accounts(); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("accounts %v", got)
	}
	if _, err := NewGroup("g2", " ", nil, time.Now()); err == nil {
		t.Fatal("empty name accepted")
	}
	if err := g.Rename("01234567890123456789012345678901234567890"); err == nil {
		t.Fatal("41 characters accepted")
	}
}

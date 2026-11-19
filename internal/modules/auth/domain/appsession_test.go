package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewDeviceName(t *testing.T) {
	if got, err := NewDeviceName("  Pixel   9 "); err != nil || got != "Pixel 9" {
		t.Fatalf("got %q, %v", got, err)
	}
	for _, bad := range []string{"", "   ", strings.Repeat("x", 81)} {
		if _, err := NewDeviceName(bad); err == nil {
			t.Fatalf("%q accepted", bad)
		}
	}
}

func TestNewAppSession(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	s, err := NewAppSession("owner", HashToken("wa_x"), "p1", "csrf", Client{}, "Pixel 9", now)
	if err != nil {
		t.Fatal(err)
	}
	if !s.IsApp() || s.Device() != "Pixel 9" || !s.ExpiresAt().Equal(now.Add(AppSessionTTL)) {
		t.Fatalf("unexpected session %+v", s)
	}
	if _, err := NewAppSession("owner", HashToken("wa_x"), "p1", "csrf", Client{}, " ", now); err == nil {
		t.Fatal("empty device accepted")
	}
	browser, _ := NewSession("owner", HashToken("x"), "p2", "csrf", Client{}, now, time.Hour)
	if browser.IsApp() {
		t.Fatal("browser session reported as app session")
	}
}

func TestAppSessionTokenPrefix(t *testing.T) {
	if !IsAppSessionToken("wa_abc") || IsAppSessionToken(APITokenPrefix+"abc") || IsAppSessionToken("abc") {
		t.Fatal("prefix not recognised")
	}
}

func TestNewPairingCode(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	c, err := NewPairingCode(HashToken("c"), "owner", now)
	if err != nil || !c.ExpiresAt.Equal(now.Add(PairingCodeTTL)) {
		t.Fatalf("got %+v, %v", c, err)
	}
	if _, err := NewPairingCode("", "owner", now); err == nil {
		t.Fatal("empty hash accepted")
	}
}

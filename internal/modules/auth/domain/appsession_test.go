package domain

import (
	"bytes"
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

func TestHandoverCodeVerifier(t *testing.T) {
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	// The example of RFC 7636, appendix B.
	c, err := NewHandoverCode(HashToken("h"), "owner", "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM", now)
	if err != nil {
		t.Fatal(err)
	}
	if !c.AcceptsVerifier("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk") || c.AcceptsVerifier("other") || c.AcceptsVerifier("") {
		t.Fatal("verifier check wrong")
	}
	if !c.ExpiresAt.Equal(now.Add(HandoverCodeTTL)) {
		t.Fatalf("expiry %v", c.ExpiresAt)
	}
	qr, _ := NewPairingCode(HashToken("q"), "owner", now)
	if !qr.AcceptsVerifier("") || qr.AcceptsVerifier("x") {
		t.Fatal("QR codes take no verifier")
	}
}

func TestNewPairingCodeText(t *testing.T) {
	// Bytes 0..7 map to the first eight characters of the alphabet; byte
	// 32 wraps around to the first one again.
	got, err := NewPairingCodeText(bytes.NewReader([]byte{0, 1, 2, 3, 4, 5, 6, 32}))
	if err != nil || got != "0123-4560" {
		t.Fatalf("got %q, %v", got, err)
	}
	if _, err := NewPairingCodeText(bytes.NewReader([]byte{1, 2})); err == nil {
		t.Fatal("short random source accepted")
	}
}

func TestCanonicalPairingCode(t *testing.T) {
	long := strings.Repeat("a", 43)
	cases := []struct{ raw, want string }{
		{"7KQ2-M9XD", "7KQ2M9XD"},
		{" 7kq2 m9xd ", "7KQ2M9XD"},
		{"oil0-abcd", "0110ABCD"},  // O, I and L read as digits
		{"7KQ2-M9X", "7KQ2-M9X"},   // too short: left alone
		{"7KQ2-M9XU", "7KQ2-M9XU"}, // U is not in the alphabet
		{" " + long + " ", long},   // handover codes are only trimmed
	}
	for _, c := range cases {
		if got := CanonicalPairingCode(c.raw); got != c.want {
			t.Errorf("CanonicalPairingCode(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}

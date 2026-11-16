package domain

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
	"time"
)

func keys(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func day(s string) time.Time {
	d, _ := time.Parse(time.DateOnly, s)
	return d
}

func sample() Licence {
	return Licence{ID: "k1", Licensee: "Example GmbH", Issued: day("2026-11-16"), Until: day("2027-11-15"), Seats: 5,
		Features: []string{FeatureBranding, FeatureSSO, "holograms"}}
}

func TestParseReadsASignedKey(t *testing.T) {
	pub, priv := keys(t)
	l, err := Parse(Sign(sample(), priv), pub)
	if err != nil {
		t.Fatal(err)
	}
	if l.Licensee != "Example GmbH" || l.Seats != 5 || !l.Until.Equal(day("2027-11-15")) || !l.Includes(FeatureSSO) {
		t.Errorf("parsed %+v", l)
	}
	// Pasted keys often carry a line break.
	if _, err := Parse(" "+Sign(sample(), priv)+"\n", pub); err != nil {
		t.Errorf("surrounding space: %v", err)
	}
}

func TestParseRefusesForgedAndBrokenKeys(t *testing.T) {
	pub, priv := keys(t)
	otherPub, _ := keys(t)
	key := Sign(sample(), priv)
	body := strings.TrimPrefix(key, Prefix)
	payload, sig, _ := strings.Cut(body, ".")

	// More seats, same signature: the payload no longer matches.
	l := sample()
	l.Seats = 500
	forgedPayload, _, _ := strings.Cut(strings.TrimPrefix(Sign(l, priv), Prefix), ".")
	for name, tc := range map[string]struct {
		key  string
		pub  ed25519.PublicKey
		want error
	}{
		"other public key":    {key, otherPub, ErrSignature},
		"changed payload":     {Prefix + forgedPayload + "." + sig, pub, ErrSignature},
		"other format prefix": {"THL2." + body, pub, ErrMalformed},
		"no signature":        {Prefix + payload, pub, ErrMalformed},
		"short signature":     {Prefix + payload + ".AAAA", pub, ErrMalformed},
		"empty":               {"", pub, ErrMalformed},
	} {
		if _, err := Parse(tc.key, tc.pub); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v, want %v", name, err, tc.want)
		}
	}
}

func TestParseRefusesSignedNonsense(t *testing.T) {
	pub, priv := keys(t)
	for name, l := range map[string]Licence{
		"no seats":          {Licensee: "X", Issued: day("2026-01-01"), Until: day("2027-01-01"), Seats: 0},
		"no licensee":       {Licensee: " ", Issued: day("2026-01-01"), Until: day("2027-01-01"), Seats: 1},
		"ends before issue": {Licensee: "X", Issued: day("2027-01-01"), Until: day("2026-01-01"), Seats: 1},
	} {
		if _, err := Parse(Sign(l, priv), pub); !errors.Is(err, ErrMalformed) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

func TestExpiryRunsToTheEndOfTheLastDay(t *testing.T) {
	l := sample()
	if l.Expired(time.Date(2027, 11, 15, 23, 59, 0, 0, time.UTC)) {
		t.Error("expired during the last day")
	}
	if !l.Expired(time.Date(2027, 11, 16, 0, 0, 0, 0, time.UTC)) {
		t.Error("still valid the day after")
	}
}

func TestLevels(t *testing.T) {
	l := sample()
	now := day("2027-01-01")
	cases := []struct {
		name     string
		lic      *Licence
		now      time.Time
		accounts int
		feature  string
		status   Status
		level    Level
	}{
		{"no key", nil, now, 1, FeatureBranding, StatusMissing, LevelNone},
		{"valid", &l, now, 5, FeatureBranding, StatusValid, LevelFull},
		{"feature not in key", &l, now, 5, FeatureAuditLog, StatusValid, LevelNone},
		{"unknown feature in key unlocks nothing", &l, now, 5, "holograms", StatusValid, LevelNone},
		{"expired keeps data readable", &l, day("2028-01-01"), 5, FeatureBranding, StatusExpired, LevelRead},
		{"more accounts than seats", &l, now, 6, FeatureBranding, StatusOverSeats, LevelRead},
	}
	for _, tc := range cases {
		status := Evaluate(tc.lic, tc.now, tc.accounts)
		if status != tc.status {
			t.Errorf("%s: status %s, want %s", tc.name, status, tc.status)
		}
		if got := LevelOf(tc.lic, status, tc.feature); got != tc.level {
			t.Errorf("%s: level %d, want %d", tc.name, got, tc.level)
		}
	}
	if got := l.UnknownFeatures(); len(got) != 1 || got[0] != "holograms" {
		t.Errorf("unknown features %v", got)
	}
}

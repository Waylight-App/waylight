package answers

import (
	"slices"
	"testing"
)

func TestFoldAndTokenize(t *testing.T) {
	if got := Fold("Büro Straße"); got != "buro strasse" {
		t.Errorf("Fold = %q", got)
	}
	got := Tokenize("Wie voll ist die Platte? (home-assistant)")
	if want := []string{"voll", "platte", "home-assistant"}; !slices.Equal(got, want) {
		t.Errorf("Tokenize = %v, want %v", got, want)
	}
}

func TestMatchToken(t *testing.T) {
	cases := []struct {
		token, term string
		want        Quality
	}{
		{"docker", "docker", Exact},
		{"do", "docker", Prefix},
		{"d", "docker", NoMatch},
		{"jelyfin", "jellyfin", Fuzzy},
		{"serien", "ferien", NoMatch}, // first letter differs
		{"abc", "abd", NoMatch},       // short words only exact or prefix
		{"festpalte", "festplatte", Fuzzy},
	}
	for _, c := range cases {
		if got := MatchToken(c.token, c.term); got != c.want {
			t.Errorf("MatchToken(%q, %q) = %v, want %v", c.token, c.term, got, c.want)
		}
	}
}

func TestBoundedDistanceStopsEarly(t *testing.T) {
	if d := boundedDistance([]rune("abcdefgh"), []rune("zyxwvuts"), 2); d != 3 {
		t.Errorf("distance = %d, want limit+1", d)
	}
	if d := boundedDistance([]rune("ab"), []rune("ba"), 1); d != 1 {
		t.Errorf("transposition = %d, want 1", d)
	}
}

func TestParseLang(t *testing.T) {
	for in, want := range map[string]Lang{"de-DE,de;q=0.9,en;q=0.8": German, "en-US": English, "fr": English, "": English, "fr,de;q=0.5": German} {
		if got := ParseLang(in); got != want {
			t.Errorf("ParseLang(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMessages(t *testing.T) {
	m := Messages{English: {"of": "{used} of {total}", "only": "en"}, German: {"of": "{used} von {total}"}}
	if got := m.T(German, "of", "used", "1 GiB", "total", 4); got != "1 GiB von 4" {
		t.Errorf("T = %q", got)
	}
	if got := m.T(German, "only"); got != "en" {
		t.Errorf("fallback = %q", got)
	}
	if got := m.T(German, "missing"); got != "missing" {
		t.Errorf("missing key = %q", got)
	}
	if FormatBytes(1536) != "1.5 KiB" || FormatBytes(20<<30) != "20 GiB" || FormatBytes(12) != "12 B" {
		t.Error("FormatBytes")
	}
}

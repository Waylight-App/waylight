package domain

import (
	"strings"
	"testing"
)

func TestNewAppearance(t *testing.T) {
	tests := []struct {
		name            string
		theme           Theme
		accent, bg, css string
		wantErr         bool
	}{
		{"defaults", ThemeSystem, "", "", "", false},
		{"full", ThemeDark, "#AABBCC", "https://example.com/bg.jpg", "body{}", false},
		{"bad theme", "pink", "", "", "", true},
		{"bad color", ThemeLight, "red", "", "", true},
		{"javascript url", ThemeLight, "", "javascript:alert(1)", "", true},
		{"css injection", ThemeLight, "", "https://x.y/a.png);background:red", "", true},
		{"huge css", ThemeLight, "", "", strings.Repeat("a", MaxCustomCSSBytes+1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAppearance(tt.theme, tt.accent, tt.bg, tt.css)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && tt.accent != "" && a.AccentColor() != strings.ToLower(tt.accent) {
				t.Fatal("accent color not normalised")
			}
		})
	}
}

func TestLanguage(t *testing.T) {
	a := Default()
	if err := a.SetLanguage(" DE "); err != nil || a.Language() != "de" {
		t.Fatalf("%q %v", a.Language(), err)
	}
	if err := a.SetLanguage("fr"); err == nil {
		t.Fatal("unsupported language accepted")
	}
}

func TestGreetingNameAndLauncher(t *testing.T) {
	a := Default()
	if a.GreetingName() != "" || a.Launcher() != LauncherBoth {
		t.Fatal("defaults")
	}
	if err := a.SetGreetingName("  Tristan "); err != nil || a.GreetingName() != "Tristan" {
		t.Fatalf("%q %v", a.GreetingName(), err)
	}
	for _, bad := range []string{strings.Repeat("x", 41), "Tris\ntan"} {
		if err := a.SetGreetingName(bad); err == nil {
			t.Errorf("name %q accepted", bad)
		}
	}
	if err := a.SetLauncher("services"); err != nil || a.Launcher() != LauncherServices {
		t.Fatalf("launcher %v", err)
	}
	if err := a.SetLauncher(""); err != nil || a.Launcher() != LauncherBoth {
		t.Fatal("empty launcher is both")
	}
	if err := a.SetLauncher("widgets"); err == nil {
		t.Fatal("unknown launcher accepted")
	}
}

// Applying a theme or a preset changes the look only: the profile, the
// greeting and its line by time of day stay (#512).
func TestStyleKeepsPersonalSettings(t *testing.T) {
	a := Default()
	if err := a.SetProfile("everyday"); err != nil {
		t.Fatal(err)
	}
	if err := a.SetGreetingName("Alex"); err != nil {
		t.Fatal(err)
	}
	a.SetGreetingMoment(true)
	styled, err := a.WithStyle("/* trisse-home theme: x */", ThemeDark, "")
	if err != nil {
		t.Fatal(err)
	}
	schemed, err := a.WithScheme(ThemeLight, "#0055aa")
	if err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string]Appearance{"style": styled, "scheme": schemed} {
		if b.Profile() != ProfileEveryday || b.GreetingName() != "Alex" || !b.GreetingMoment() {
			t.Errorf("%s: %+v", name, b)
		}
	}
}

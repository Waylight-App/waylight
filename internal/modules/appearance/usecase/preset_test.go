package usecase

import (
	"context"
	"testing"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/shared/kernel"
)

type presetRepo struct {
	a   domain.Appearance
	own bool
}

func (m *presetRepo) Load(context.Context) (domain.Appearance, error) { return m.a, nil }
func (m *presetRepo) Save(_ context.Context, a domain.Appearance) error {
	m.a, m.own = a, true
	return nil
}

func as(role string) context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "x", Role: role})
}

func TestPresetForAccountsWithoutTheirOwn(t *testing.T) {
	repo := &presetRepo{a: domain.Default()}
	p := NewPresets(func(context.Context) (bool, error) { return repo.own, nil })
	p.SetSource(func(context.Context) (Preset, bool) { return Preset{Theme: domain.ThemeDark, Accent: "#0055aa"}, true })
	get, update := NewGetAppearance(repo, p), NewUpdateAppearance(repo, p)

	res, _ := get.Execute(as("member"))
	if res.Theme != "dark" || res.AccentColor != "#0055aa" || res.ThemeLocked {
		t.Fatalf("new account: %+v", res)
	}
	// Not locked: the member's own choice wins once saved.
	if res, err := update.Execute(as("member"), UpdateAppearanceRequest{Theme: "light", Launcher: "both", Profile: "everyday"}); err != nil || res.Theme != "light" {
		t.Fatalf("own choice: %+v %v", res, err)
	}
	if res, _ := get.Execute(as("member")); res.Theme != "light" {
		t.Fatalf("after saving: %+v", res)
	}
}

func TestLockedPresetKeepsMembersButNotAdmins(t *testing.T) {
	repo := &presetRepo{a: domain.Default(), own: true}
	p := NewPresets(func(context.Context) (bool, error) { return repo.own, nil })
	p.SetSource(func(context.Context) (Preset, bool) {
		return Preset{Theme: domain.ThemeDark, Accent: "#0055aa", Locked: true}, true
	})
	get, update := NewGetAppearance(repo, p), NewUpdateAppearance(repo, p)

	res, err := update.Execute(as("member"), UpdateAppearanceRequest{Theme: "light", AccentColor: "#ff0000", Language: "de", Launcher: "both", Profile: "everyday"})
	if err != nil || res.Theme != "dark" || res.AccentColor != "#0055aa" || res.Language != "de" || !res.ThemeLocked {
		t.Fatalf("member: %+v %v", res, err)
	}
	if res, _ := get.Execute(as("member")); res.Theme != "dark" || !res.ThemeLocked {
		t.Fatalf("member get: %+v", res)
	}
	if res, err := update.Execute(as("admin"), UpdateAppearanceRequest{Theme: "light", Launcher: "both", Profile: "everyday"}); err != nil ||
		res.Theme != "light" || res.ThemeLocked {
		t.Fatalf("admin: %+v %v", res, err)
	}
}

func defaultThemes(t *testing.T) []domain.ThemeDef {
	t.Helper()
	colors := map[string]string{"--color-bg": "#000", "--color-surface": "#111", "--color-text": "#fff", "--color-accent": "#f0f"}
	single, err := domain.NewThemeDef("night", "Night", "", domain.ThemeDark, colors, nil)
	if err != nil {
		t.Fatal(err)
	}
	both, err := domain.NewThemeDef("both", "Both", "", "", nil, map[domain.Theme]map[string]string{domain.ThemeLight: colors, domain.ThemeDark: colors})
	if err != nil {
		t.Fatal(err)
	}
	return []domain.ThemeDef{single, both}
}

// DEFAULT_THEME (#507): accounts without their own appearance get it, an
// own appearance wins, and an organisation's preset goes on top.
func TestDefaultTheme(t *testing.T) {
	repo := &presetRepo{a: domain.Default()}
	p := NewPresets(func(context.Context) (bool, error) { return repo.own, nil })
	if err := p.SetDefaultTheme(defaultThemes(t), "nope"); err == nil {
		t.Fatal("unknown id accepted")
	}
	if _, ok := p.DefaultTheme(); ok {
		t.Fatal("an unknown id must not set a default")
	}
	if err := p.SetDefaultTheme(defaultThemes(t), "night"); err != nil {
		t.Fatal(err)
	}
	get, update := NewGetAppearance(repo, p), NewUpdateAppearance(repo, p)

	res, _ := get.Execute(as("member"))
	if domain.AppliedTheme(res.CustomCSS) != "night" || res.Theme != "dark" {
		t.Fatalf("new account: %+v", res)
	}
	// A theme with variants keeps the stored light/dark/system choice.
	_ = p.SetDefaultTheme(defaultThemes(t), "both")
	if res, _ := get.Execute(as("member")); domain.AppliedTheme(res.CustomCSS) != "both" || res.Theme != "system" {
		t.Fatalf("variants: %+v", res)
	}
	// The organisation's scheme and accent go on top.
	p.SetSource(func(context.Context) (Preset, bool) { return Preset{Theme: domain.ThemeLight, Accent: "#0055aa"}, true })
	if res, _ := get.Execute(as("member")); domain.AppliedTheme(res.CustomCSS) != "both" || res.Theme != "light" || res.AccentColor != "#0055aa" {
		t.Fatalf("with preset: %+v", res)
	}
	// Once the account saves its own appearance, the default is gone.
	if _, err := update.Execute(as("member"), UpdateAppearanceRequest{Theme: "dark", Launcher: "both", Profile: "everyday"}); err != nil {
		t.Fatal(err)
	}
	if res, _ := get.Execute(as("member")); res.CustomCSS != "" || res.Theme != "dark" {
		t.Fatalf("own: %+v", res)
	}
	// Empty keeps the built-in look.
	if err := p.SetDefaultTheme(nil, ""); err != nil {
		t.Fatal(err)
	}
	if _, ok := p.DefaultTheme(); ok {
		t.Fatal("cleared default still set")
	}
}

func TestNoPresetChangesNothing(t *testing.T) {
	repo := &presetRepo{a: domain.Default()}
	res, _ := NewGetAppearance(repo, nil).Execute(as("member"))
	if res.Theme != "system" || res.ThemeLocked {
		t.Fatalf("%+v", res)
	}
}

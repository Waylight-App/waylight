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

func TestNoPresetChangesNothing(t *testing.T) {
	repo := &presetRepo{a: domain.Default()}
	res, _ := NewGetAppearance(repo, nil).Execute(as("member"))
	if res.Theme != "system" || res.ThemeLocked {
		t.Fatalf("%+v", res)
	}
}

package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/appearance/domain"
	"waylight/internal/shared/kernel"
)

type memRepo struct{ a *domain.Appearance }

func (m *memRepo) Load(context.Context) (domain.Appearance, error) {
	if m.a == nil {
		return domain.Default(), nil
	}
	return *m.a, nil
}

func (m *memRepo) Save(_ context.Context, a domain.Appearance) error { m.a = &a; return nil }

func TestAppearanceUseCases(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	if got, _ := NewGetAppearance(repo, nil).Execute(ctx); got.Theme != "system" {
		t.Fatalf("default theme = %q", got.Theme)
	}
	if _, err := NewUpdateAppearance(repo, nil).Execute(ctx, UpdateAppearanceRequest{Theme: "neon"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("expected invalid, got %v", err)
	}
	if _, err := NewUpdateAppearance(repo, nil).Execute(ctx, UpdateAppearanceRequest{Theme: "dark", AccentColor: "#112233"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := NewGetAppearance(repo, nil).Execute(ctx); got.Theme != "dark" || got.AccentColor != "#112233" {
		t.Fatalf("not saved: %+v", got)
	}
}

// TestShowAdvancedKeptWhenNotSent: a client that does not know the switch
// (#602) saves its form without turning the advanced settings off.
func TestShowAdvancedKeptWhenNotSent(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	on, off := true, false
	update := NewUpdateAppearance(repo, nil)
	if got, err := update.Execute(ctx, UpdateAppearanceRequest{Theme: "system", ShowAdvanced: &on}); err != nil || !got.ShowAdvanced {
		t.Fatalf("switch on: %+v %v", got, err)
	}
	if got, err := update.Execute(ctx, UpdateAppearanceRequest{Theme: "dark"}); err != nil || !got.ShowAdvanced {
		t.Fatalf("an old client turned it off: %+v %v", got, err)
	}
	if got, err := update.Execute(ctx, UpdateAppearanceRequest{Theme: "dark", ShowAdvanced: &off}); err != nil || got.ShowAdvanced {
		t.Fatalf("switch off: %+v %v", got, err)
	}
	if got, _ := NewGetAppearance(repo, nil).Execute(ctx); got.ShowAdvanced {
		t.Fatal("not stored")
	}
}

// TestAccessibilityKeptWhenNotSent: the options of #681 are stored, kept by
// clients that do not send them, and a wrong text size is refused.
func TestAccessibilityKeptWhenNotSent(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	on := true
	update := NewUpdateAppearance(repo, nil)
	got, err := update.Execute(ctx, UpdateAppearanceRequest{Theme: "system", TextSize: "larger", HighContrast: &on, SimpleView: &on})
	if err != nil || got.TextSize != "larger" || !got.HighContrast || !got.SimpleView {
		t.Fatalf("set: %+v %v", got, err)
	}
	if got, err := update.Execute(ctx, UpdateAppearanceRequest{Theme: "dark"}); err != nil || got.TextSize != "larger" || !got.HighContrast || !got.SimpleView {
		t.Fatalf("an old client reset them: %+v %v", got, err)
	}
	if _, err := update.Execute(ctx, UpdateAppearanceRequest{Theme: "dark", TextSize: "huge"}); err == nil {
		t.Fatal("unknown text size accepted")
	}
	if got, _ := NewGetAppearance(repo, nil).Execute(ctx); got.TextSize != "larger" || !got.SimpleView {
		t.Fatalf("not stored: %+v", got)
	}
}

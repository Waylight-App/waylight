package usecase

import (
	"context"
	"errors"
	"slices"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/shared/kernel"
)

// Preset is an organisation's colour scheme (#305): the theme and accent
// colour an account without its own appearance starts with, and whether
// members may change them.
type Preset struct {
	Theme  domain.Theme
	Accent string
	// Locked keeps members on the preset; admins can still change theirs.
	Locked bool
}

// Presets are the defaults an account starts with before it stores its own
// appearance: the instance's default theme (DEFAULT_THEME, #507) and the
// organisation's colour scheme from the branding (ee/, #305). The core does
// not import ee/: the composition root sets the sources.
type Presets struct {
	source func(ctx context.Context) (Preset, bool)
	// owned tells whether the principal has stored its own appearance.
	owned func(ctx context.Context) (bool, error)
	// defaultTheme is the instance's default theme, or nil for the
	// built-in look.
	defaultTheme *domain.ThemeDef
}

// NewPresets returns presets without a source: nothing changes.
func NewPresets(owned func(ctx context.Context) (bool, error)) *Presets {
	return &Presets{owned: owned}
}

// SetSource sets where the preset comes from.
func (p *Presets) SetSource(src func(ctx context.Context) (Preset, bool)) { p.source = src }

// ErrUnknownDefaultTheme is returned for a DEFAULT_THEME that names no
// built-in theme.
var ErrUnknownDefaultTheme = errors.New("DEFAULT_THEME names no built-in theme")

// SetDefaultTheme makes a built-in theme the default for the sign-in page
// and for every account without its own appearance. An empty id keeps the
// built-in look.
func (p *Presets) SetDefaultTheme(builtin []domain.ThemeDef, id string) error {
	if id == "" {
		p.defaultTheme = nil
		return nil
	}
	i := slices.IndexFunc(builtin, func(t domain.ThemeDef) bool { return t.ID == id })
	if i < 0 {
		return ErrUnknownDefaultTheme
	}
	t := builtin[i]
	p.defaultTheme = &t
	return nil
}

// DefaultTheme returns the instance's default theme, if one is set.
func (p *Presets) DefaultTheme() (domain.ThemeDef, bool) {
	if p == nil || p.defaultTheme == nil {
		return domain.ThemeDef{}, false
	}
	return *p.defaultTheme, true
}

func (p *Presets) current(ctx context.Context) (Preset, bool) {
	if p == nil || p.source == nil {
		return Preset{}, false
	}
	pr, ok := p.source(ctx)
	if !ok || pr.Theme == "" {
		return Preset{}, false
	}
	return pr, true
}

// lockedFor reports whether the preset binds the principal: members,
// guests and kiosk displays, but not admins.
func lockedFor(ctx context.Context, pr Preset) bool {
	return pr.Locked && !kernel.PrincipalFrom(ctx).IsAdmin()
}

// apply puts the defaults on an appearance loaded for the principal. An
// account's own appearance always wins over the default theme. The
// organisation's preset applies to an account without its own appearance,
// or to everyone it binds, and its scheme and accent colour go on top of the
// default theme.
func (p *Presets) apply(ctx context.Context, a domain.Appearance) (domain.Appearance, bool, error) {
	if p == nil {
		return a, false, nil
	}
	pr, hasPreset := p.current(ctx)
	locked := hasPreset && lockedFor(ctx, pr)
	own := false
	if p.defaultTheme != nil || (hasPreset && !locked) {
		var err error
		if own, err = p.owned(ctx); err != nil {
			return a, false, err
		}
	}
	if own && !locked {
		return a, false, nil
	}
	if p.defaultTheme != nil && !own {
		// A single-scheme theme brings its scheme; one with variants keeps
		// the stored choice. The theme brings its own accent colour.
		if b, err := a.WithStyle(p.defaultTheme.CSS(), p.defaultTheme.Scheme, ""); err == nil {
			a = b
		}
	}
	if !hasPreset {
		return a, false, nil
	}
	b, err := a.WithScheme(pr.Theme, pr.Accent)
	if err != nil {
		return a, locked, nil // a broken preset never locks anyone out
	}
	return b, locked, nil
}

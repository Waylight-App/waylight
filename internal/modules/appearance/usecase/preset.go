package usecase

import (
	"context"

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

// Presets connects the organisation's branding (ee/) to the appearance
// without the core importing ee/: the composition root sets the source.
type Presets struct {
	source func(ctx context.Context) (Preset, bool)
	// owned tells whether the principal has stored its own appearance.
	owned func(ctx context.Context) (bool, error)
}

// NewPresets returns presets without a source: nothing changes.
func NewPresets(owned func(ctx context.Context) (bool, error)) *Presets {
	return &Presets{owned: owned}
}

// SetSource sets where the preset comes from.
func (p *Presets) SetSource(src func(ctx context.Context) (Preset, bool)) { p.source = src }

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

// apply puts the preset on an appearance loaded for the principal: for an
// account without its own appearance, or for everyone it binds.
func (p *Presets) apply(ctx context.Context, a domain.Appearance) (domain.Appearance, bool, error) {
	pr, ok := p.current(ctx)
	if !ok {
		return a, false, nil
	}
	locked := lockedFor(ctx, pr)
	if !locked {
		own, err := p.owned(ctx)
		if err != nil {
			return a, false, err
		}
		if own {
			return a, false, nil
		}
	}
	b, err := a.WithScheme(pr.Theme, pr.Accent)
	if err != nil {
		return a, locked, nil // a broken preset never locks anyone out
	}
	return b, locked, nil
}

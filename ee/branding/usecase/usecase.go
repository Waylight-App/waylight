// Package usecase changes and supplies an organisation's branding (#305).
//
// Ring: Use cases.
package usecase

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"time"

	"trisse-home/ee/branding/domain"
	licence "trisse-home/ee/licence/domain"
	"trisse-home/internal/platform/branding"
	"trisse-home/internal/shared/imaging"
	"trisse-home/internal/shared/kernel"
)

// Asset is a stored image.
type Asset struct {
	Data        []byte
	ContentType string
	UpdatedAt   time.Time
}

// Store keeps the settings and images (output port).
type Store interface {
	Settings(ctx context.Context) (domain.Settings, error)
	SaveSettings(ctx context.Context, s domain.Settings) error
	Asset(ctx context.Context, kind string) (Asset, bool, error)
	// Assets returns when each stored image was last changed.
	Assets(ctx context.Context) (map[string]time.Time, error)
	SaveAsset(ctx context.Context, kind string, a Asset) error
	DeleteAsset(ctx context.Context, kind string) error
}

// Licensing tells whether branding is licensed (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

// Deps are the collaborators.
type Deps struct {
	Store     Store
	Licensing Licensing
	Clock     kernel.Clock
	Audit     func(ctx context.Context, action, id string)
}

// View is the settings page.
type View struct {
	domain.Settings
	Assets map[string]time.Time
	// Applied is false while the licence does not cover branding; the stored
	// branding is kept and comes back with a valid key.
	Applied  bool
	Warnings []string
}

// Service implements the use cases and branding.Source.
type Service struct{ d Deps }

// New returns the use cases.
func New(d Deps) *Service { return &Service{d: d} }

func admin(ctx context.Context) error {
	if p := kernel.PrincipalFrom(ctx); !p.IsAdmin() || p.IsToken() {
		return domain.ErrNotAllowed
	}
	return nil
}

func (s *Service) applies(ctx context.Context) bool {
	return s.d.Licensing.Level(ctx, licence.FeatureBranding) == licence.LevelFull
}

// Get returns the settings page.
func (s *Service) Get(ctx context.Context) (View, error) {
	if err := admin(ctx); err != nil {
		return View{}, err
	}
	set, err := s.d.Store.Settings(ctx)
	if err != nil {
		return View{}, err
	}
	assets, err := s.d.Store.Assets(ctx)
	if err != nil {
		return View{}, err
	}
	return View{Settings: set, Assets: assets, Applied: s.applies(ctx), Warnings: set.Warnings()}, nil
}

// Save changes name, colours and the product credit.
func (s *Service) Save(ctx context.Context, set domain.Settings) (View, error) {
	if err := admin(ctx); err != nil {
		return View{}, err
	}
	if err := set.Validate(); err != nil {
		return View{}, err
	}
	if err := s.d.Store.SaveSettings(ctx, set); err != nil {
		return View{}, fmt.Errorf("save branding: %w", err)
	}
	s.audit(ctx, "settings", "branding")
	return s.Get(ctx)
}

func validKind(kind string) bool {
	return kind == branding.Logo || kind == branding.LogoDark || kind == branding.Favicon
}

// Upload stores an image. It is decoded and re-encoded as PNG, scaled to
// at most 512 pixels: nothing of the uploaded file reaches a browser as it
// was sent, which rules out scripts hidden in it.
func (s *Service) Upload(ctx context.Context, kind string, data []byte) (View, error) {
	if err := admin(ctx); err != nil {
		return View{}, err
	}
	if !validKind(kind) {
		return View{}, domain.ErrKind
	}
	if len(data) == 0 || len(data) > domain.MaxUpload {
		return View{}, domain.ErrImage
	}
	img, err := imaging.Decode(data)
	if err != nil {
		// Unsupported formats (SVG, WebP), too many pixels and broken files
		// all get the same answer.
		return View{}, domain.ErrImage
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, imaging.Fit(img, domain.MaxEdge)); err != nil {
		return View{}, fmt.Errorf("encode image: %w", err)
	}
	if err := s.d.Store.SaveAsset(ctx, kind, Asset{Data: buf.Bytes(), ContentType: "image/png", UpdatedAt: s.d.Clock.Now()}); err != nil {
		return View{}, fmt.Errorf("save image: %w", err)
	}
	s.audit(ctx, "upload", kind)
	return s.Get(ctx)
}

// Delete removes an image; the product's is shown again.
func (s *Service) Delete(ctx context.Context, kind string) (View, error) {
	if err := admin(ctx); err != nil {
		return View{}, err
	}
	if !validKind(kind) {
		return View{}, domain.ErrKind
	}
	if err := s.d.Store.DeleteAsset(ctx, kind); err != nil {
		return View{}, err
	}
	s.audit(ctx, "delete", kind)
	return s.Get(ctx)
}

// Branding implements branding.Source: the organisation's branding while
// the licence covers it and something is set.
func (s *Service) Branding(ctx context.Context) (branding.Branding, bool) {
	if !s.applies(ctx) {
		return branding.Branding{}, false
	}
	set, err := s.d.Store.Settings(ctx)
	if err != nil {
		return branding.Branding{}, false
	}
	assets, err := s.d.Store.Assets(ctx)
	if err != nil {
		return branding.Branding{}, false
	}
	if set.Empty() && len(assets) == 0 {
		return branding.Branding{}, false
	}
	return branding.Branding{Name: set.Name, Accent: set.Accent, Secondary: set.Secondary, HidePoweredBy: set.HidePoweredBy,
		DefaultTheme: set.DefaultTheme, LockTheme: set.LockTheme, Assets: assets}, true
}

// Asset implements branding.Source.
func (s *Service) Asset(ctx context.Context, kind string) ([]byte, string, bool) {
	a, ok, err := s.d.Store.Asset(ctx, kind)
	if err != nil || !ok {
		return nil, "", false
	}
	return a.Data, a.ContentType, true
}

func (s *Service) audit(ctx context.Context, action, id string) {
	if s.d.Audit != nil {
		s.d.Audit(ctx, action, id)
	}
}

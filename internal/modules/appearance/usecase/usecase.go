package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/appearance/domain"
)

// AppearanceResponse is the use case output.
type AppearanceResponse struct {
	Theme           string
	Language        string
	AccentColor     string
	BackgroundImage string
	CustomCSS       string
	GreetingName    string
	GreetingMoment  bool
	Launcher        string
	Profile         string
	ShowAdvanced    bool
	// ThemeLocked: the organisation set theme and accent colour (#305).
	ThemeLocked bool
}

// UpdateAppearanceRequest is the input of UpdateAppearance.
type UpdateAppearanceRequest struct {
	Theme           string
	Language        string
	AccentColor     string
	BackgroundImage string
	CustomCSS       string
	GreetingName    string
	GreetingMoment  bool
	Launcher        string
	// Profile is "everyday" or "homelab" (#281); empty keeps the stored one.
	Profile string
	// ShowAdvanced lists advanced settings (#602); nil keeps the stored value,
	// so clients that do not know the switch never turn it off.
	ShowAdvanced *bool
}

// GetAppearance returns the current preferences (input port).
type GetAppearance interface {
	Execute(ctx context.Context) (AppearanceResponse, error)
}

// UpdateAppearance validates and stores new preferences (input port).
type UpdateAppearance interface {
	Execute(ctx context.Context, req UpdateAppearanceRequest) (AppearanceResponse, error)
}

// Repository persists the appearance (output port). Load returns
// domain.Default() when nothing has been stored yet.
type Repository interface {
	Load(ctx context.Context) (domain.Appearance, error)
	Save(ctx context.Context, a domain.Appearance) error
}

func toResponse(a domain.Appearance) AppearanceResponse {
	return AppearanceResponse{Theme: string(a.Theme()), AccentColor: a.AccentColor(), BackgroundImage: a.BackgroundImage(), CustomCSS: a.CustomCSS(), Language: a.Language(),
		GreetingName: a.GreetingName(), GreetingMoment: a.GreetingMoment(), Launcher: string(a.Launcher()), Profile: string(a.Profile()),
		ShowAdvanced: a.ShowAdvanced()}
}

type getAppearance struct {
	repo    Repository
	presets *Presets
}

// NewGetAppearance returns the GetAppearance use case; presets may be nil.
func NewGetAppearance(repo Repository, presets *Presets) GetAppearance {
	return &getAppearance{repo: repo, presets: presets}
}

// Execute implements GetAppearance.
func (uc *getAppearance) Execute(ctx context.Context) (AppearanceResponse, error) {
	a, err := uc.repo.Load(ctx)
	if err != nil {
		return AppearanceResponse{}, fmt.Errorf("load appearance: %w", err)
	}
	a, locked, err := uc.presets.apply(ctx, a)
	if err != nil {
		return AppearanceResponse{}, fmt.Errorf("apply preset: %w", err)
	}
	res := toResponse(a)
	res.ThemeLocked = locked
	return res, nil
}

type updateAppearance struct {
	repo    Repository
	presets *Presets
}

// NewUpdateAppearance returns the UpdateAppearance use case; presets may be
// nil.
func NewUpdateAppearance(repo Repository, presets *Presets) UpdateAppearance {
	return &updateAppearance{repo: repo, presets: presets}
}

// Execute implements UpdateAppearance.
func (uc *updateAppearance) Execute(ctx context.Context, req UpdateAppearanceRequest) (AppearanceResponse, error) {
	// A locked preset keeps theme and accent colour; the rest is saved.
	pr, ok := uc.presets.current(ctx)
	locked := ok && lockedFor(ctx, pr)
	if locked {
		req.Theme, req.AccentColor = string(pr.Theme), pr.Accent
	}
	a, err := domain.NewAppearance(domain.Theme(req.Theme), req.AccentColor, req.BackgroundImage, req.CustomCSS)
	if err == nil {
		err = a.SetLanguage(req.Language)
	}
	if err == nil {
		err = a.SetGreetingName(req.GreetingName)
	}
	a.SetGreetingMoment(req.GreetingMoment)
	if err == nil {
		err = a.SetLauncher(req.Launcher)
	}
	if err == nil && (req.Profile == "" || req.ShowAdvanced == nil) {
		// Older clients send neither field; both then keep the stored value.
		stored, lerr := uc.repo.Load(ctx)
		if lerr != nil {
			return AppearanceResponse{}, fmt.Errorf("load appearance: %w", lerr)
		}
		if req.Profile == "" {
			req.Profile = string(stored.Profile())
		}
		if req.ShowAdvanced == nil {
			keep := stored.ShowAdvanced()
			req.ShowAdvanced = &keep
		}
	}
	if err == nil {
		err = a.SetProfile(req.Profile)
	}
	if req.ShowAdvanced != nil {
		a.SetShowAdvanced(*req.ShowAdvanced)
	}
	if err != nil {
		return AppearanceResponse{}, err
	}
	if err := uc.repo.Save(ctx, a); err != nil {
		return AppearanceResponse{}, fmt.Errorf("save appearance: %w", err)
	}
	res := toResponse(a)
	res.ThemeLocked = locked
	return res, nil
}

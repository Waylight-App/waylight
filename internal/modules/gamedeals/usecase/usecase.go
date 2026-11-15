package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"trisse-home/internal/modules/gamedeals/domain"
	"trisse-home/internal/shared/kernel"
)

// MaxDeals bounds the wishlist deals shown.
const MaxDeals = 15

// GameDeals is the input port of the module (#62).
type GameDeals interface {
	Settings(ctx context.Context) (domain.Settings, error)
	SaveSettings(ctx context.Context, in domain.Settings) (domain.Settings, error)
	Overview(ctx context.Context) (domain.Overview, error)
	// Refresh reloads both sources (job, daily).
	Refresh(ctx context.Context) error
}

// SettingsRepository persists the settings (output port).
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// WishlistProvider reads the discounted games of a public wishlist.
type WishlistProvider interface {
	Wishlist(ctx context.Context, steamID, country string) ([]domain.Deal, error)
}

// FreeGamesProvider reads the free game promotions.
type FreeGamesProvider interface {
	FreeGames(ctx context.Context, country string) ([]domain.FreeGame, error)
}

// Deps bundles the output ports.
type Deps struct {
	Repo     SettingsRepository
	Wishlist WishlistProvider
	Free     FreeGamesProvider
	Clock    kernel.Clock
}

type gameDeals struct {
	d  Deps
	mu sync.RWMutex
	// cache holds the offers per set of sources (ADR 0018): accounts with
	// the same Steam ID, country and Epic choice share one entry, so all
	// accounts that use the admin's settings share the admin's.
	cache map[sourceKey]*cached
}

// sourceKey identifies what is fetched; the minimum discount only filters.
type sourceKey struct {
	steamID, country string
	epic             bool
}

type cached struct {
	settings domain.Settings
	deals    []domain.Deal
	free     []domain.FreeGame
	errors   []string
}

func keyOf(s domain.Settings) sourceKey { return sourceKey{s.SteamID, s.Country, s.Epic} }

// NewGameDeals returns the use cases.
func NewGameDeals(d Deps) GameDeals { return &gameDeals{d: d, cache: map[sourceKey]*cached{}} }

// Settings implements GameDeals.
func (uc *gameDeals) Settings(ctx context.Context) (domain.Settings, error) {
	return uc.d.Repo.Load(ctx)
}

// SaveSettings implements GameDeals; the sources are loaded right away so
// a private wishlist is reported immediately.
func (uc *gameDeals) SaveSettings(ctx context.Context, in domain.Settings) (domain.Settings, error) {
	s, err := domain.NewSettings(in.SteamID, in.Country, in.MinDiscount, in.Epic)
	if err != nil {
		return domain.Settings{}, err
	}
	if s.SteamID != "" {
		if _, err := uc.d.Wishlist.Wishlist(ctx, s.SteamID, s.Country); err != nil {
			return domain.Settings{}, kernel.NewPublicErrorf(kernel.ErrInvalid, "Could not load the Steam wishlist (is it public?): {detail}", "detail", err.Error())
		}
	}
	if err := uc.d.Repo.Save(ctx, s); err != nil {
		return domain.Settings{}, fmt.Errorf("save settings: %w", err)
	}
	_ = uc.refreshFor(ctx, s) // errors are shown in the overview
	return s, nil
}

// Overview implements GameDeals (cache-aside per set of sources).
func (uc *gameDeals) Overview(ctx context.Context) (domain.Overview, error) {
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return domain.Overview{}, err
	}
	uc.mu.RLock()
	_, loaded := uc.cache[keyOf(s)]
	uc.mu.RUnlock()
	if !loaded {
		_ = uc.refreshFor(ctx, s)
	}
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	c := uc.cache[keyOf(s)]
	if c == nil {
		return domain.Overview{}, nil // sources switched off
	}
	now := uc.d.Clock.Now()
	o := domain.Overview{Errors: slices.Clone(c.errors), Deals: domain.FilterDeals(c.deals, s.MinDiscount)}
	o.Deals = o.Deals[:min(len(o.Deals), MaxDeals)]
	for _, g := range c.free {
		switch {
		case g.Current(now):
			o.FreeNow = append(o.FreeNow, g)
		case g.Start.After(now):
			o.FreeNext = append(o.FreeNext, g)
		}
	}
	return o, nil
}

// Refresh implements GameDeals: the job reloads the principal's sources
// (the admin's, when it runs as the system) and every set of sources that
// was asked for since the start.
func (uc *gameDeals) Refresh(ctx context.Context) error {
	s, err := uc.d.Repo.Load(ctx)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}
	uc.mu.RLock()
	all := []domain.Settings{s}
	for k, c := range uc.cache {
		if k != keyOf(s) {
			all = append(all, c.settings)
		}
	}
	uc.mu.RUnlock()
	var errs []error
	for _, one := range all {
		errs = append(errs, uc.refreshFor(ctx, one))
	}
	return errors.Join(errs...)
}

// refreshFor loads one set of sources; each source fails on its own and
// keeps its previous offers.
func (uc *gameDeals) refreshFor(ctx context.Context, s domain.Settings) error {
	if !s.Enabled() {
		return nil
	}
	var deals []domain.Deal
	var free []domain.FreeGame
	var failed []string
	var errs []error
	var err error
	if s.SteamID != "" {
		if deals, err = uc.d.Wishlist.Wishlist(ctx, s.SteamID, s.Country); err != nil {
			failed, errs = append(failed, "Steam"), append(errs, err)
		}
	}
	if s.Epic {
		if free, err = uc.d.Free.FreeGames(ctx, s.Country); err != nil {
			failed, errs = append(failed, "Epic Games Store"), append(errs, err)
		}
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	c := uc.cache[keyOf(s)]
	if c == nil {
		c = &cached{}
		uc.cache[keyOf(s)] = c
	}
	c.settings = s
	if !slices.Contains(failed, "Steam") {
		c.deals = deals
	}
	if !slices.Contains(failed, "Epic Games Store") {
		c.free = free
	}
	c.errors = failed
	return errors.Join(errs...)
}

// RefreshInterval is how often the job runs: offers change at most daily.
const RefreshInterval = 12 * time.Hour

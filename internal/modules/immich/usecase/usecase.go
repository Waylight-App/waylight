package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"waylight/internal/modules/immich/domain"
	"waylight/internal/shared/kernel"
)

// overviewTTL bounds how often Immich is asked; statistics change slowly
// and are fetched on demand only.
const overviewTTL = 5 * time.Minute

// SettingsView never contains the API key.
type SettingsView struct {
	Configured bool
	URL        string
	HasAPIKey  bool
}

// SettingsInput is the input of SaveSettings; an empty key keeps the stored one.
type SettingsInput struct {
	URL, APIKey string
}

// Immich is the input port of the module (#85).
type Immich interface {
	Settings(ctx context.Context) (SettingsView, error)
	SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error)
	// Overview reads the library through the named connection the widget
	// chose (#610), or through the module settings when connectionID is
	// empty or the connection is not usable.
	Overview(ctx context.Context, connectionID string) (domain.Overview, error)
	// Thumbnail returns the preview of a photo of today's memories.
	Thumbnail(ctx context.Context, connectionID, assetID string) (domain.Thumbnail, error)
}

// Connections resolves a named connection of the integrations module to
// settings (output port, #610). Any error means the module settings are
// used instead.
type Connections interface {
	Settings(ctx context.Context, id string) (domain.Settings, error)
}

// SettingsRepository persists the settings (output port). Load returns
// domain.ErrNotConfigured when nothing is stored.
type SettingsRepository interface {
	Load(ctx context.Context) (domain.Settings, error)
	Save(ctx context.Context, s domain.Settings) error
}

// Client talks to the Immich API (output port).
type Client interface {
	Version(ctx context.Context, s domain.Settings) (string, error)
	Stats(ctx context.Context, s domain.Settings) (domain.LibraryStats, error)
	// Memories returns today's memories; an older Immich without the
	// endpoint returns none.
	Memories(ctx context.Context, s domain.Settings, day time.Time) ([]domain.Memory, error)
	Thumbnail(ctx context.Context, s domain.Settings, assetID string) (domain.Thumbnail, error)
}

type cachedOverview struct {
	overview domain.Overview
	fetched  time.Time
}

type immich struct {
	repo   SettingsRepository
	client Client
	conns  Connections
	clock  kernel.Clock
	mu     sync.Mutex
	// cache holds one overview per connection ID; "" is the module
	// settings. There are at most as many entries as connections.
	cache map[string]cachedOverview
}

// NewImmich returns the use cases; conns may be nil, then only the module
// settings are used.
func NewImmich(repo SettingsRepository, client Client, conns Connections, clock kernel.Clock) Immich {
	return &immich{repo: repo, client: client, conns: conns, clock: clock, cache: map[string]cachedOverview{}}
}

// load returns the settings of the named connection, falling back to the
// module settings while the connection does not exist or is not usable,
// so a widget keeps working whatever happened to the copy (#610).
func (uc *immich) load(ctx context.Context, connectionID string) (domain.Settings, error) {
	if connectionID != "" && uc.conns != nil {
		if s, err := uc.conns.Settings(ctx, connectionID); err == nil {
			return s, nil
		}
	}
	return uc.repo.Load(ctx)
}

// Settings implements Immich.
func (uc *immich) Settings(ctx context.Context) (SettingsView, error) {
	s, err := uc.repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return SettingsView{}, nil
	}
	if err != nil {
		return SettingsView{}, fmt.Errorf("load settings: %w", err)
	}
	return SettingsView{Configured: true, URL: s.URL, HasAPIKey: true}, nil
}

// SaveSettings implements Immich; the key is tested with the statistics
// call, which needs the asset.statistics permission.
func (uc *immich) SaveSettings(ctx context.Context, in SettingsInput) (SettingsView, error) {
	key := in.APIKey
	if key == "" {
		if old, err := uc.repo.Load(ctx); err == nil {
			key = old.APIKey
		}
	}
	s, err := domain.NewSettings(in.URL, key)
	if err != nil {
		return SettingsView{}, err
	}
	if _, err := uc.client.Stats(ctx, s); err != nil {
		return SettingsView{}, kernel.SetupFailure("Immich", err)
	}
	if err := uc.repo.Save(ctx, s); err != nil {
		return SettingsView{}, fmt.Errorf("save settings: %w", err)
	}
	uc.mu.Lock()
	clear(uc.cache)
	uc.mu.Unlock()
	return SettingsView{Configured: true, URL: s.URL, HasAPIKey: true}, nil
}

// Overview implements Immich. Memories and version are optional: their
// errors (e.g. a key without memory.read) only leave them empty.
func (uc *immich) Overview(ctx context.Context, connectionID string) (domain.Overview, error) {
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return domain.Overview{}, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	now := uc.clock.Now()
	if c, ok := uc.cache[connectionID]; ok && now.Sub(c.fetched) < overviewTTL && sameDay(c.fetched, now) {
		return c.overview, nil
	}
	stats, err := uc.client.Stats(ctx, s)
	if err != nil {
		return domain.Overview{}, kernel.NewPublicError(kernel.ErrUnavailable, "Immich is not reachable.")
	}
	o := domain.Overview{Stats: stats}
	o.Version, _ = uc.client.Version(ctx, s)       // optional
	memories, _ := uc.client.Memories(ctx, s, now) // optional, needs memory.read
	for _, m := range memories {
		if len(m.AssetIDs) > domain.MaxMemoryAssets {
			m.AssetIDs = m.AssetIDs[:domain.MaxMemoryAssets]
		}
		o.Memories = append(o.Memories, m)
	}
	uc.cache[connectionID] = cachedOverview{overview: o, fetched: now}
	return o, nil
}

// Thumbnail implements Immich. Only assets of today's memories are
// proxied, so the endpoint cannot be used to browse the library.
func (uc *immich) Thumbnail(ctx context.Context, connectionID, assetID string) (domain.Thumbnail, error) {
	if !domain.ValidAssetID(assetID) {
		return domain.Thumbnail{}, domain.ErrThumbnailNotFound
	}
	o, err := uc.Overview(ctx, connectionID)
	if err != nil {
		return domain.Thumbnail{}, err
	}
	known := slices.ContainsFunc(o.Memories, func(m domain.Memory) bool { return slices.Contains(m.AssetIDs, assetID) })
	if !known {
		return domain.Thumbnail{}, domain.ErrThumbnailNotFound
	}
	s, err := uc.load(ctx, connectionID)
	if err != nil {
		return domain.Thumbnail{}, err
	}
	t, err := uc.client.Thumbnail(ctx, s, assetID)
	if err != nil {
		return domain.Thumbnail{}, domain.ErrThumbnailNotFound
	}
	return t, nil
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

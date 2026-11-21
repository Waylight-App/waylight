package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/immich/domain"
	"waylight/internal/shared/kernel"
)

type memRepo struct{ s *domain.Settings }

func (m *memRepo) Load(context.Context) (domain.Settings, error) {
	if m.s == nil {
		return domain.Settings{}, domain.ErrNotConfigured
	}
	return *m.s, nil
}
func (m *memRepo) Save(_ context.Context, s domain.Settings) error { m.s = &s; return nil }

const asset = "3f2b6c1e-0a4d-4e5f-9a8b-1c2d3e4f5a6b"

type fakeImmich struct {
	statsCalls int
	noMemories bool
	thumbs     []string
}

func (f *fakeImmich) Version(context.Context, domain.Settings) (string, error) { return "1.120.0", nil }
func (f *fakeImmich) Stats(_ context.Context, s domain.Settings) (domain.LibraryStats, error) {
	if s.APIKey != "key" {
		return domain.LibraryStats{}, errors.New("the API key was rejected")
	}
	f.statsCalls++
	return domain.LibraryStats{Photos: 100, Videos: 5}, nil
}
func (f *fakeImmich) Memories(context.Context, domain.Settings, time.Time) ([]domain.Memory, error) {
	if f.noMemories {
		return nil, errors.New("the API key lacks a permission")
	}
	ids := make([]string, domain.MaxMemoryAssets+2)
	for i := range ids {
		ids[i] = asset
	}
	return []domain.Memory{{YearsAgo: 3, AssetIDs: ids}}, nil
}
func (f *fakeImmich) Thumbnail(_ context.Context, _ domain.Settings, id string) (domain.Thumbnail, error) {
	f.thumbs = append(f.thumbs, id)
	return domain.Thumbnail{ContentType: "image/webp", Data: []byte{1}}, nil
}

type movingClock struct{ now time.Time }

func (c *movingClock) Now() time.Time { return c.now }

func TestImmich(t *testing.T) {
	ctx := context.Background()
	repo, client := &memRepo{}, &fakeImmich{}
	c := &movingClock{now: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)}
	uc := NewImmich(repo, client, nil, c)

	if _, err := uc.Overview(ctx, ""); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://immich", APIKey: "bad"}); !errors.Is(err, kernel.ErrInvalid) || repo.s != nil {
		t.Fatalf("a failing connection test must not save: %v", err)
	}
	if _, err := uc.SaveSettings(ctx, SettingsInput{URL: "http://immich", APIKey: "key"}); err != nil {
		t.Fatal(err)
	}
	o, err := uc.Overview(ctx, "")
	if err != nil || o.Stats.Photos != 100 || o.Version != "1.120.0" || len(o.Memories[0].AssetIDs) != domain.MaxMemoryAssets {
		t.Fatalf("overview: %+v %v", o, err)
	}
	calls := client.statsCalls
	_, _ = uc.Overview(ctx, "")
	if client.statsCalls != calls {
		t.Fatal("cached within the TTL")
	}
	c.now = c.now.Add(14 * time.Hour) // next day: memories change
	_, _ = uc.Overview(ctx, "")
	if client.statsCalls != calls+1 {
		t.Fatal("a new day loads new memories")
	}

	if _, err := uc.Thumbnail(ctx, "", asset); err != nil || len(client.thumbs) != 1 {
		t.Fatalf("thumbnail of a memory: %v", err)
	}
	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "../api/users"} {
		if _, err := uc.Thumbnail(ctx, "", id); !errors.Is(err, domain.ErrThumbnailNotFound) {
			t.Errorf("%q: only memory assets are proxied, got %v", id, err)
		}
	}

	client.noMemories = true
	c.now = c.now.Add(overviewTTL)
	if o, err := uc.Overview(ctx, ""); err != nil || len(o.Memories) != 0 {
		t.Fatalf("a key without memory.read still shows statistics: %+v %v", o, err)
	}
}

// conns resolves "office" to another Immich; everything else fails, like a
// missing copy or a connection of another service.
type conns struct{}

func (conns) Settings(_ context.Context, id string) (domain.Settings, error) {
	if id == "office" {
		return domain.Settings{URL: "http://office", APIKey: "key"}, nil
	}
	return domain.Settings{}, errors.New("no usable connection")
}

// recordingClient remembers which Immich was asked.
type recordingClient struct {
	fakeImmich
	urls []string
}

func (r *recordingClient) Stats(ctx context.Context, s domain.Settings) (domain.LibraryStats, error) {
	r.urls = append(r.urls, s.URL)
	return r.fakeImmich.Stats(ctx, s)
}

func TestOverviewThroughConnection(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{s: &domain.Settings{URL: "http://home", APIKey: "key"}}
	client := &recordingClient{}
	uc := NewImmich(repo, client, conns{}, &movingClock{now: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC)})
	for _, id := range []string{"office", "", "missing"} {
		if _, err := uc.Overview(ctx, id); err != nil {
			t.Fatalf("%q: %v", id, err)
		}
	}
	// "missing" falls back to the module settings, cached on its own key.
	if want := []string{"http://office", "http://home", "http://home"}; strings.Join(client.urls, " ") != strings.Join(want, " ") {
		t.Fatalf("asked %v, want %v", client.urls, want)
	}
	_, _ = uc.Overview(ctx, "office")
	if len(client.urls) != 3 {
		t.Fatal("each connection is cached")
	}
	repo.s = nil
	if _, err := uc.Overview(ctx, "office"); err != nil {
		t.Fatalf("a connection works without module settings: %v", err)
	}
}

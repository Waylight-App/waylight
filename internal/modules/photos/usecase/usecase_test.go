package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"trisse-home/internal/modules/photos/domain"
	"trisse-home/internal/shared/kernel"
)

type memRepo struct {
	photos []domain.Photo
	usage  domain.Usage
}

func (m *memRepo) List(_ context.Context, album string) ([]domain.Photo, error) {
	var out []domain.Photo
	for _, p := range m.photos {
		if p.Album == album {
			out = append(out, p)
		}
	}
	return out, nil
}
func (m *memRepo) Albums(context.Context) ([]domain.Album, error) { return nil, nil }
func (m *memRepo) Usage(context.Context) (domain.Usage, error)    { return m.usage, nil }
func (m *memRepo) Add(_ context.Context, p domain.Photo, _, _ []byte) (domain.Photo, error) {
	m.photos = append(m.photos, p)
	m.usage.Count++
	m.usage.Bytes += p.Size
	return p, nil
}
func (m *memRepo) Delete(context.Context, string) error { return nil }
func (m *memRepo) Reorder(_ context.Context, _ string, ids []string) error {
	slices.SortFunc(m.photos, func(a, b domain.Photo) int { return slices.Index(ids, a.ID) - slices.Index(ids, b.ID) })
	return nil
}
func (m *memRepo) Image(context.Context, string, bool) ([]byte, error) { return nil, nil }

type proc struct{ fail error }

func (p proc) Process(data []byte) (Processed, error) {
	if p.fail != nil {
		return Processed{}, p.fail
	}
	return Processed{Image: data, Thumb: []byte("t"), Width: 4, Height: 3}, nil
}

type clk struct{}

func (clk) Now() time.Time { return time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC) }

type seq struct{ n int }

func (s *seq) NewID() string { s.n++; return string(rune('a' + s.n - 1)) }

func TestUpload(t *testing.T) {
	ctx := context.Background()
	r := &memRepo{}
	uc := NewPhotos(Deps{Repo: r, Processor: proc{}, Clock: clk{}, IDs: &seq{}})
	p, err := uc.Upload(ctx, "", []byte("12345"))
	if err != nil || p.Album != domain.DefaultAlbum || p.Size != 6 || p.Width != 4 {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err := uc.Upload(ctx, "x", make([]byte, domain.MaxUploadBytes+1)); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("too large: %v", err)
	}
	uc = NewPhotos(Deps{Repo: r, Processor: proc{fail: ErrUnsupportedImage}, Clock: clk{}, IDs: &seq{}})
	if _, err := uc.Upload(ctx, "x", []byte("svg")); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("unsupported: %v", err)
	}
	r.usage = domain.Usage{Bytes: domain.QuotaBytes - 3}
	uc = NewPhotos(Deps{Repo: r, Processor: proc{}, Clock: clk{}, IDs: &seq{}})
	if _, err := uc.Upload(ctx, "x", []byte("12345")); !errors.Is(err, kernel.ErrConflict) {
		t.Fatalf("quota: %v", err)
	}
}

func TestReorder(t *testing.T) {
	ctx := context.Background()
	r := &memRepo{photos: []domain.Photo{{ID: "a", Album: "A"}, {ID: "b", Album: "A"}, {ID: "c", Album: "B"}}}
	uc := NewPhotos(Deps{Repo: r})
	if err := uc.Reorder(ctx, "A", []string{"b", "a"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := uc.List(ctx, "A"); got[0].ID != "b" {
		t.Fatalf("%+v", got)
	}
	if err := uc.Reorder(ctx, "A", []string{"a", "c"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("foreign picture accepted: %v", err)
	}
}

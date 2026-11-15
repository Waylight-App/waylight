package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/shared/kernel"
)

type memBackgrounds struct{ list []domain.Background }

func (m *memBackgrounds) ListBackgrounds(context.Context) ([]domain.Background, error) {
	return m.list, nil
}
func (m *memBackgrounds) AddBackground(_ context.Context, b domain.Background, _, _ []byte) error {
	m.list = append(m.list, b)
	return nil
}
func (m *memBackgrounds) BackgroundImage(context.Context, string, bool) ([]byte, error) {
	return nil, nil
}
func (m *memBackgrounds) DeleteBackground(_ context.Context, id string) error {
	for i, b := range m.list {
		if b.ID == id {
			m.list = append(m.list[:i], m.list[i+1:]...)
			return nil
		}
	}
	return domain.ErrBackgroundNotFound
}

type fakeProcessor struct{ fail error }

func (p fakeProcessor) Process(data []byte, _, _ int) (ProcessedImage, error) {
	return ProcessedImage{Image: data, Thumb: []byte("t"), Width: 4, Height: 3}, p.fail
}

type bgClock struct{}

func (bgClock) Now() time.Time { return time.Unix(0, 0) }

type bgIDs struct{ n int }

func (i *bgIDs) NewID() string { i.n++; return string(rune('a' + i.n - 1)) }

func TestBackgrounds(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	store := &memBackgrounds{}
	uc := NewBackgrounds(BackgroundDeps{Store: store, Processor: fakeProcessor{}, Repo: repo, Clock: bgClock{}, IDs: &bgIDs{}})
	b, err := uc.Upload(ctx, []byte("jpeg"))
	if err != nil || b.ID != "a" || b.Size != 5 {
		t.Fatalf("%+v %v", b, err)
	}
	res, err := uc.Set(ctx, b.URL())
	if err != nil || res.BackgroundImage != "/api/appearance/backgrounds/a" {
		t.Fatalf("set: %+v %v", res, err)
	}
	if _, err := uc.Set(ctx, "javascript:alert(1)"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("invalid value: %v", err)
	}
	// Deleting the picture in use falls back to the default background.
	if err := uc.Delete(ctx, "a"); err != nil || repo.a.BackgroundImage() != "" {
		t.Fatalf("delete: %v %q", err, repo.a.BackgroundImage())
	}
	uc = NewBackgrounds(BackgroundDeps{Store: store, Processor: fakeProcessor{fail: ErrUnsupportedImage}, Repo: repo, Clock: bgClock{}, IDs: &bgIDs{}})
	if _, err := uc.Upload(ctx, []byte("<svg/>")); !errors.Is(err, domain.ErrBackgroundUnsupported) {
		t.Fatalf("unsupported: %v", err)
	}
	store.list = make([]domain.Background, domain.MaxBackgrounds)
	if _, err := uc.Upload(ctx, []byte("x")); !errors.Is(err, kernel.ErrConflict) {
		t.Fatalf("limit: %v", err)
	}
}

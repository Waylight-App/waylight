package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/places/domain"
)

type memRepo struct{ m map[string]domain.Place }

func (r *memRepo) List(context.Context) ([]domain.Place, error) {
	out := make([]domain.Place, 0, len(r.m))
	for _, p := range r.m {
		out = append(out, p)
	}
	return out, nil
}

func (r *memRepo) Get(_ context.Context, id string) (domain.Place, error) {
	p, ok := r.m[id]
	if !ok {
		return domain.Place{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *memRepo) Save(_ context.Context, p domain.Place) error { r.m[p.ID] = p; return nil }
func (r *memRepo) Delete(_ context.Context, id string) error    { delete(r.m, id); return nil }

type clock struct{}

func (clock) Now() time.Time { return time.Unix(1_800_000_000, 0) }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return string(rune('a' + i.n - 1)) }

func newUC(widgets []Widget) (Places, *memRepo) {
	repo := &memRepo{m: map[string]domain.Place{}}
	return NewPlaces(Deps{
		Repo: repo, Clock: clock{}, IDs: &ids{},
		Widgets: func(context.Context) ([]Widget, error) { return widgets, nil },
	}), repo
}

func TestPlacesNameTheWidgetsThatUseThem(t *testing.T) {
	uc, _ := newUC([]Widget{
		{ID: "w1", TabID: "t1", Kind: "weather", PlaceID: "a"},
		{ID: "w2", TabID: "t1", Kind: "transit"},
	})
	ctx := context.Background()
	home, err := uc.Add(ctx, Input{Name: "Home", Lat: 53.14, Lon: 8.21})
	if err != nil || home.ID != "a" {
		t.Fatalf("%+v %v", home, err)
	}
	if _, err := uc.Add(ctx, Input{Name: "Office", Lat: 52.5, Lon: 13.4}); err != nil {
		t.Fatal(err)
	}
	list, _ := uc.List(ctx)
	for _, p := range list {
		if p.ID == "a" && (len(p.UsedBy) != 1 || p.UsedBy[0].WidgetID != "w1" || p.UsedBy[0].Kind != "weather") {
			t.Fatalf("home: %+v", p)
		}
		if p.ID == "b" && len(p.UsedBy) != 0 {
			t.Fatalf("office: %+v", p)
		}
	}

	moved, err := uc.Update(ctx, "a", Input{Name: "Home", Label: "Bremen", Lat: 53.08, Lon: 8.8})
	if err != nil || moved.Label != "Bremen" || len(moved.UsedBy) != 1 {
		t.Fatalf("%+v %v", moved, err)
	}
	if err := uc.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, "a"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestAddStopsAtTheLimit(t *testing.T) {
	uc, repo := newUC(nil)
	for i := range domain.MaxPlaces {
		repo.m[string(rune(1000+i))] = domain.Place{}
	}
	if _, err := uc.Add(context.Background(), Input{Name: "One more"}); !errors.Is(err, domain.ErrTooMany) {
		t.Fatalf("got %v", err)
	}
}

package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/places/domain"
	"waylight/internal/shared/kernel"
)

// PlaceView is a place with the widgets that take their location from it.
type PlaceView struct {
	domain.Place
	UsedBy []domain.Use
}

// Input is what a person enters for a place.
type Input struct {
	Name, Label string
	Lat, Lon    float64
}

// Places is the input port of the module (#591); every method is one use case.
type Places interface {
	List(ctx context.Context) ([]PlaceView, error)
	Add(ctx context.Context, in Input) (PlaceView, error)
	Update(ctx context.Context, id string, in Input) (PlaceView, error)
	Delete(ctx context.Context, id string) error
}

// Repository persists places (output port, implemented in adapter/gateway).
// It sees only the places visible to the principal in ctx and changes only
// the principal's own (ADR 0018).
type Repository interface {
	List(ctx context.Context) ([]domain.Place, error)
	Get(ctx context.Context, id string) (domain.Place, error)
	Save(ctx context.Context, p domain.Place) error
	Delete(ctx context.Context, id string) error
}

// Widget is a widget on one of the account's own tabs with the place its
// settings name ("placeId"), empty for none.
type Widget struct {
	ID, TabID, Kind, PlaceID string
}

// Deps bundles the output ports.
type Deps struct {
	Repo Repository
	// Widgets lists the widgets on the account's own tabs; nil when the
	// dashboard is not available, then no place is in use.
	Widgets func(ctx context.Context) ([]Widget, error)
	Clock   kernel.Clock
	IDs     kernel.IDGenerator
}

type uc struct{ d Deps }

// NewPlaces returns the use cases.
func NewPlaces(d Deps) Places { return &uc{d: d} }

// List implements Places.
func (u *uc) List(ctx context.Context) ([]PlaceView, error) {
	places, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list places: %w", err)
	}
	uses, err := u.uses(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PlaceView, 0, len(places))
	for _, p := range places {
		out = append(out, PlaceView{Place: p, UsedBy: uses[p.ID]})
	}
	return out, nil
}

// Add implements Places.
func (u *uc) Add(ctx context.Context, in Input) (PlaceView, error) {
	existing, err := u.d.Repo.List(ctx)
	if err != nil {
		return PlaceView{}, fmt.Errorf("list places: %w", err)
	}
	if len(existing) >= domain.MaxPlaces {
		return PlaceView{}, domain.ErrTooMany
	}
	p, err := domain.NewPlace(u.d.IDs.NewID(), in.Name, in.Label, in.Lat, in.Lon, u.d.Clock.Now())
	if err != nil {
		return PlaceView{}, err
	}
	if err := u.d.Repo.Save(ctx, p); err != nil {
		return PlaceView{}, fmt.Errorf("save place: %w", err)
	}
	return PlaceView{Place: p}, nil
}

// Update implements Places.
func (u *uc) Update(ctx context.Context, id string, in Input) (PlaceView, error) {
	p, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return PlaceView{}, err
	}
	if err := p.Change(in.Name, in.Label, in.Lat, in.Lon); err != nil {
		return PlaceView{}, err
	}
	if err := u.d.Repo.Save(ctx, p); err != nil {
		return PlaceView{}, err
	}
	uses, err := u.uses(ctx)
	if err != nil {
		return PlaceView{}, err
	}
	return PlaceView{Place: p, UsedBy: uses[p.ID]}, nil
}

// Delete implements Places. Widgets that used the place keep the
// coordinates they copied from it, so they go on working; the client asks
// before it deletes a place in use.
func (u *uc) Delete(ctx context.Context, id string) error {
	if _, err := u.d.Repo.Get(ctx, id); err != nil {
		return err
	}
	return u.d.Repo.Delete(ctx, id)
}

// uses maps place IDs to the widgets whose settings name them.
func (u *uc) uses(ctx context.Context) (map[string][]domain.Use, error) {
	out := map[string][]domain.Use{}
	if u.d.Widgets == nil {
		return out, nil
	}
	widgets, err := u.d.Widgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list widgets: %w", err)
	}
	for _, w := range widgets {
		if w.PlaceID != "" {
			out[w.PlaceID] = append(out[w.PlaceID], domain.Use{WidgetID: w.ID, TabID: w.TabID, Kind: w.Kind})
		}
	}
	return out, nil
}

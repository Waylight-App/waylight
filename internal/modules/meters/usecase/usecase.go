package usecase

import (
	"context"
	"fmt"
	"slices"
	"time"

	"waylight/internal/modules/meters/domain"
	"waylight/internal/shared/kernel"
)

// Meters is the input port of the module; every method is one use case.
type Meters interface {
	// List returns the visible meters by name, each with its last reading.
	List(ctx context.Context) ([]MeterView, error)
	// Create adds a meter; an empty visibility shares it with everyone,
	// since meters belong to the household.
	Create(ctx context.Context, in domain.Input, visibility string) (MeterView, error)
	// Update changes a meter of the principal; an empty visibility keeps it.
	Update(ctx context.Context, id string, in domain.Input, visibility string) (MeterView, error)
	// Delete removes a meter of the principal with its readings.
	Delete(ctx context.Context, id string) error
	// Detail returns the readings of a visible meter, newest first, and the
	// consumption of year (0 = the current year).
	Detail(ctx context.Context, id string, year int) (Detail, error)
	// AddReading stores a reading of a visible meter; a reading on a day
	// that has one replaces it.
	AddReading(ctx context.Context, id string, in domain.ReadingInput) (domain.Reading, error)
	// DeleteReading removes the reading of a day from a meter of the principal.
	DeleteReading(ctx context.Context, id, day string) error
}

// MeterView is a meter with its latest reading.
type MeterView struct {
	domain.Meter
	// Last is nil while the meter has no reading.
	Last *domain.Reading
	// Mine says whether the principal created the meter.
	Mine bool
}

// Detail is a meter with its readings and the consumption of one year.
type Detail struct {
	MeterView
	Readings    []domain.Reading
	Consumption domain.YearConsumption
	// Years are the calendar years the readings touch, the latest first.
	Years []int
}

// Repository persists meters and readings (output port, implemented in
// adapter/gateway). Reads return only what the principal in ctx may see;
// changes to a meter and removing readings need its owner, adding a reading
// only visibility (ADR 0018).
type Repository interface {
	List(ctx context.Context) ([]domain.Meter, error)
	Get(ctx context.Context, id string) (domain.Meter, error)
	// Save upserts a meter: new ones belong to the principal, existing ones
	// change only for the owner, otherwise kernel.ErrNotOwner.
	Save(ctx context.Context, m domain.Meter) error
	// Delete removes the meter and its readings; an unknown one is no error,
	// one of someone else is kernel.ErrNotOwner.
	Delete(ctx context.Context, id string) error
	// LastReadings returns the newest reading of each visible meter.
	LastReadings(ctx context.Context) (map[string]domain.Reading, error)
	// Readings returns the readings of a visible meter by day, oldest first.
	Readings(ctx context.Context, id string) ([]domain.Reading, error)
	// SaveReading stores a reading of a visible meter, replacing the one of
	// the same day; domain.ErrNotFound for any other meter.
	SaveReading(ctx context.Context, id string, r domain.Reading) error
	// DeleteReading removes a reading of a meter of the principal.
	DeleteReading(ctx context.Context, id string, day time.Time) error
}

// Deps bundles the output ports. Location decides what "today" is
// (nil = time.Local).
type Deps struct {
	Repo     Repository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct{ d Deps }

// NewMeters returns the meter use cases.
func NewMeters(d Deps) Meters {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

func (u *uc) today() time.Time {
	n := u.d.Clock.Now().In(u.d.Location)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

func (u *uc) view(ctx context.Context, m domain.Meter, last map[string]domain.Reading) MeterView {
	v := MeterView{Meter: m, Mine: m.Access.Mine(kernel.PrincipalFrom(ctx))}
	if r, ok := last[m.ID]; ok {
		v.Last = &r
	}
	return v
}

// List implements Meters.
func (u *uc) List(ctx context.Context) ([]MeterView, error) {
	list, err := u.d.Repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list meters: %w", err)
	}
	last, err := u.d.Repo.LastReadings(ctx)
	if err != nil {
		return nil, fmt.Errorf("last readings: %w", err)
	}
	out := make([]MeterView, 0, len(list))
	for _, m := range list {
		out = append(out, u.view(ctx, m, last))
	}
	return out, nil
}

// Create implements Meters.
func (u *uc) Create(ctx context.Context, in domain.Input, visibility string) (MeterView, error) {
	p := kernel.PrincipalFrom(ctx)
	access := kernel.Access{Owner: p.AccountID, Visibility: kernel.Everyone}
	if visibility != "" {
		v, err := kernel.CheckShare(p, visibility)
		if err != nil {
			return MeterView{}, err
		}
		access.Visibility = v
	}
	m, err := domain.New(u.d.IDs.NewID(), in, access, u.d.Clock.Now())
	if err != nil {
		return MeterView{}, err
	}
	if err := u.d.Repo.Save(ctx, m); err != nil {
		return MeterView{}, fmt.Errorf("save meter: %w", err)
	}
	return u.view(ctx, m, nil), nil
}

// own loads a meter the principal created.
func (u *uc) own(ctx context.Context, id string) (domain.Meter, error) {
	m, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return domain.Meter{}, err
	}
	if !m.Access.Mine(kernel.PrincipalFrom(ctx)) {
		return domain.Meter{}, kernel.ErrNotOwner
	}
	return m, nil
}

// Update implements Meters.
func (u *uc) Update(ctx context.Context, id string, in domain.Input, visibility string) (MeterView, error) {
	m, err := u.own(ctx, id)
	if err != nil {
		return MeterView{}, err
	}
	if visibility != "" {
		if m.Access.Visibility, err = kernel.CheckShare(kernel.PrincipalFrom(ctx), visibility); err != nil {
			return MeterView{}, err
		}
	}
	if m, err = m.Update(in); err != nil {
		return MeterView{}, err
	}
	if err := u.d.Repo.Save(ctx, m); err != nil {
		return MeterView{}, err
	}
	last, err := u.d.Repo.LastReadings(ctx)
	if err != nil {
		return MeterView{}, err
	}
	return u.view(ctx, m, last), nil
}

// Delete implements Meters.
func (u *uc) Delete(ctx context.Context, id string) error { return u.d.Repo.Delete(ctx, id) }

// Detail implements Meters.
func (u *uc) Detail(ctx context.Context, id string, year int) (Detail, error) {
	m, err := u.d.Repo.Get(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	readings, err := u.d.Repo.Readings(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	if year == 0 {
		year = u.today().Year()
	}
	d := Detail{
		MeterView:   u.view(ctx, m, nil),
		Readings:    slices.Clone(readings),
		Consumption: domain.ConsumptionOf(readings, year),
		Years:       domain.Years(readings),
	}
	slices.Reverse(d.Readings)
	if n := len(readings); n > 0 {
		d.Last = &readings[n-1]
	}
	return d, nil
}

// AddReading implements Meters.
func (u *uc) AddReading(ctx context.Context, id string, in domain.ReadingInput) (domain.Reading, error) {
	r, err := domain.NewReading(in, u.today())
	if err != nil {
		return domain.Reading{}, err
	}
	if _, err := u.d.Repo.Get(ctx, id); err != nil {
		return domain.Reading{}, err
	}
	existing, err := u.d.Repo.Readings(ctx, id)
	if err != nil {
		return domain.Reading{}, err
	}
	if err := domain.CheckOrder(existing, r); err != nil {
		return domain.Reading{}, err
	}
	if err := u.d.Repo.SaveReading(ctx, id, r); err != nil {
		return domain.Reading{}, fmt.Errorf("save reading: %w", err)
	}
	return r, nil
}

// DeleteReading implements Meters.
func (u *uc) DeleteReading(ctx context.Context, id, day string) error {
	if _, err := u.own(ctx, id); err != nil {
		return err
	}
	d, err := domain.ParseDay("day", day)
	if err != nil {
		return err
	}
	return u.d.Repo.DeleteReading(ctx, id, d)
}

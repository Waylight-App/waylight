// Package usecase sets and lists live presence (#346).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"waylight/ee/presence/domain"
	"waylight/internal/shared/kernel"
)

// Store keeps today's statuses and the on/off switch (output port).
type Store interface {
	Entries(ctx context.Context) ([]domain.Entry, error)
	Save(ctx context.Context, e domain.Entry) error
	Delete(ctx context.Context, accountID string) error
	// DeleteBefore removes statuses of days before day: no history.
	DeleteBefore(ctx context.Context, day string) error
	Enabled(ctx context.Context) (bool, error)
	SetEnabled(ctx context.Context, on bool) error
}

// Directory knows the people of the instance and its groups (output port,
// implemented over the auth module in cmd/server).
type Directory interface {
	// People returns account ID → name of everyone who can sign in
	// (owner, admins, members; not guests or kiosk displays).
	People(ctx context.Context) (map[string]string, error)
	// GroupMembers returns the account IDs of a group; domain.ErrNoGroup
	// for an unknown group.
	GroupMembers(ctx context.Context, groupID string) ([]string, error)
}

// Absences reports planned absences of #299 (output port); nil without
// the absences module.
type Absences interface {
	AwayOn(ctx context.Context, day string) (map[string]domain.Absent, error)
}

// Deps are the collaborators.
type Deps struct {
	Store     Store
	Directory Directory
	Absences  Absences
	Clock     kernel.Clock
	// Location decides when a day ends; time.Local by default.
	Location *time.Location
}

// Board is the presence of a group today.
type Board struct {
	Day    string
	People []domain.Person
	Counts map[domain.Status]int
	// Mine is the viewer's own entry as shown (StatusNone without one).
	Mine domain.Person
}

// Service implements the use cases.
type Service struct{ d Deps }

// New returns the use cases.
func New(d Deps) *Service {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &Service{d: d}
}

func (s *Service) today() string { return s.d.Clock.Now().In(s.d.Location).Format(domain.DateLayout) }

func (s *Service) checkEnabled(ctx context.Context) error {
	on, err := s.d.Store.Enabled(ctx)
	if err != nil {
		return fmt.Errorf("read switch: %w", err)
	}
	if !on {
		return domain.ErrDisabled
	}
	return nil
}

// Board lists a group's members with today's status; an empty groupID
// lists everyone who can sign in.
func (s *Service) Board(ctx context.Context, groupID string) (Board, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return Board{}, err
	}
	people, err := s.d.Directory.People(ctx)
	if err != nil {
		return Board{}, fmt.Errorf("list people: %w", err)
	}
	ids := make([]string, 0, len(people))
	if groupID == "" {
		for id := range people {
			ids = append(ids, id)
		}
	} else {
		if ids, err = s.d.Directory.GroupMembers(ctx, groupID); err != nil {
			return Board{}, err
		}
	}
	day := s.today()
	entries, err := s.d.Store.Entries(ctx)
	if err != nil {
		return Board{}, fmt.Errorf("list statuses: %w", err)
	}
	byAccount := make(map[string]domain.Entry, len(entries))
	for _, e := range entries {
		byAccount[e.AccountID] = e
	}
	away := map[string]domain.Absent{}
	if s.d.Absences != nil {
		// Planned absences are a bonus: without them the manual status shows.
		if a, err := s.d.Absences.AwayOn(ctx, day); err == nil {
			away = a
		}
	}
	resolve := func(id string) domain.Person {
		var e *domain.Entry
		if x, ok := byAccount[id]; ok {
			e = &x
		}
		var a *domain.Absent
		if x, ok := away[id]; ok {
			a = &x
		}
		return domain.Resolve(id, people[id], e, a, day)
	}
	b := Board{Day: day, People: make([]domain.Person, 0, len(ids))}
	for _, id := range ids {
		if _, ok := people[id]; ok {
			b.People = append(b.People, resolve(id))
		}
	}
	domain.SortPeople(b.People)
	b.Counts = domain.Counts(b.People)
	if p := kernel.PrincipalFrom(ctx); domain.MaySet(p) {
		b.Mine = resolve(p.AccountID)
	}
	return b, nil
}

// Set sets the viewer's own status for today; StatusNone clears it.
func (s *Service) Set(ctx context.Context, status domain.Status, returnOn string) error {
	p := kernel.PrincipalFrom(ctx)
	if !domain.MaySet(p) {
		return domain.ErrNotSelf
	}
	if err := s.checkEnabled(ctx); err != nil {
		return err
	}
	if status == domain.StatusNone {
		return s.d.Store.Delete(ctx, p.AccountID)
	}
	e, err := domain.NewEntry(p.AccountID, status, returnOn, s.today())
	if err != nil {
		return err
	}
	if err := s.d.Store.Save(ctx, e); err != nil {
		return fmt.Errorf("save status: %w", err)
	}
	return nil
}

// Enabled reports the switch.
func (s *Service) Enabled(ctx context.Context) (bool, error) { return s.d.Store.Enabled(ctx) }

// SetEnabled switches presence on or off for everyone (admins). Switching
// off also forgets today's statuses.
func (s *Service) SetEnabled(ctx context.Context, on bool) error {
	p := kernel.PrincipalFrom(ctx)
	if !p.IsAdmin() || p.TabScoped() || p.IsToken() {
		return domain.ErrAdmin
	}
	if err := s.d.Store.SetEnabled(ctx, on); err != nil {
		return fmt.Errorf("save switch: %w", err)
	}
	if !on {
		return s.d.Store.DeleteBefore(ctx, "9999-12-31")
	}
	return nil
}

// Prune deletes statuses of earlier days (job).
func (s *Service) Prune(ctx context.Context) error {
	if err := s.d.Store.DeleteBefore(ctx, s.today()); err != nil {
		return fmt.Errorf("prune statuses: %w", err)
	}
	return nil
}

// ForgetAccount removes a deleted account's status.
func (s *Service) ForgetAccount(ctx context.Context, accountID string) error {
	if err := s.d.Store.Delete(ctx, accountID); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("delete status: %w", err)
	}
	return nil
}

// Package usecase enters and lists absences (#299).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"fmt"
	"sort"

	"trisse-home/ee/absences/domain"
	"trisse-home/internal/shared/kernel"
)

// Store keeps absences (output port).
type Store interface {
	// Between returns the absences that touch the days from..to.
	Between(ctx context.Context, from, to string) ([]domain.Absence, error)
	Get(ctx context.Context, id string) (domain.Absence, error)
	Save(ctx context.Context, a domain.Absence) error
	Delete(ctx context.Context, id string) error
	DeleteAccount(ctx context.Context, accountID string) error
}

// Names resolves account names; "" for an unknown account (output port).
type Names interface {
	Name(ctx context.Context, accountID string) string
}

// Deps are the collaborators.
type Deps struct {
	Store Store
	Names Names
	IDs   kernel.IDGenerator
	Clock kernel.Clock
	Audit func(ctx context.Context, action, id string)
}

// View is an absence as the viewer may see it.
type View struct {
	domain.Absence
	Name     string
	Editable bool
}

// Input is what the client sends; AccountID empty means oneself.
type Input struct {
	AccountID string
	From, To  string
	Kind      domain.Kind
	Note      string
	Private   bool
}

// Service implements the use cases.
type Service struct{ d Deps }

// New returns the use cases.
func New(d Deps) *Service { return &Service{d: d} }

// List returns the absences touching from..to, at most MaxRange days, in
// order of first day and name.
func (s *Service) List(ctx context.Context, from, to string) ([]View, error) {
	f, err := domain.ParseDay(from)
	if err != nil {
		return nil, kernel.NewValidationError("from", "Enter the first day as YYYY-MM-DD.")
	}
	t, err := domain.ParseDay(to)
	if err != nil || t.Before(f) {
		return nil, kernel.NewValidationError("to", "Enter a last day as YYYY-MM-DD, not before the first.")
	}
	if t.Sub(f).Hours()/24 >= domain.MaxRange {
		return nil, kernel.NewValidationError("to", "Ask for at most 62 days at once.")
	}
	list, err := s.d.Store.Between(ctx, f.Format(domain.DateLayout), t.Format(domain.DateLayout))
	if err != nil {
		return nil, fmt.Errorf("list absences: %w", err)
	}
	p := kernel.PrincipalFrom(ctx)
	names := map[string]string{}
	out := make([]View, 0, len(list))
	for _, a := range list {
		name, ok := names[a.AccountID]
		if !ok {
			name = s.d.Names.Name(ctx, a.AccountID)
			names[a.AccountID] = name
		}
		out = append(out, View{Absence: a.SeenBy(p), Name: name, Editable: domain.MayChange(p, a.AccountID)})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// Create enters an absence for oneself or, as an admin, for anyone.
func (s *Service) Create(ctx context.Context, in Input) (View, error) {
	p := kernel.PrincipalFrom(ctx)
	account := in.AccountID
	if account == "" {
		account = p.AccountID
	}
	if !domain.MayChange(p, account) {
		return View{}, domain.ErrNotAllowed
	}
	name := s.d.Names.Name(ctx, account)
	if name == "" {
		return View{}, domain.ErrNoAccount
	}
	a := domain.Absence{ID: s.d.IDs.NewID(), AccountID: account, From: in.From, To: in.To, Kind: in.Kind, Note: in.Note,
		Private: in.Private, CreatedBy: p.AccountID, CreatedAt: s.d.Clock.Now()}
	if err := a.Validate(); err != nil {
		return View{}, err
	}
	if err := s.d.Store.Save(ctx, a); err != nil {
		return View{}, fmt.Errorf("save absence: %w", err)
	}
	s.audit(ctx, "create", a.ID)
	return View{Absence: a, Name: name, Editable: true}, nil
}

// Update changes days, kind, note and privacy; the person stays.
func (s *Service) Update(ctx context.Context, id string, in Input) (View, error) {
	a, err := s.changeable(ctx, id)
	if err != nil {
		return View{}, err
	}
	a.From, a.To, a.Kind, a.Note, a.Private = in.From, in.To, in.Kind, in.Note, in.Private
	if err := a.Validate(); err != nil {
		return View{}, err
	}
	if err := s.d.Store.Save(ctx, a); err != nil {
		return View{}, fmt.Errorf("save absence: %w", err)
	}
	s.audit(ctx, "update", a.ID)
	return View{Absence: a, Name: s.d.Names.Name(ctx, a.AccountID), Editable: true}, nil
}

// Delete removes an absence.
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.changeable(ctx, id); err != nil {
		return err
	}
	if err := s.d.Store.Delete(ctx, id); err != nil {
		return err
	}
	s.audit(ctx, "delete", id)
	return nil
}

// ForgetAccount removes a deleted account's absences.
func (s *Service) ForgetAccount(ctx context.Context, accountID string) error {
	return s.d.Store.DeleteAccount(ctx, accountID)
}

func (s *Service) changeable(ctx context.Context, id string) (domain.Absence, error) {
	a, err := s.d.Store.Get(ctx, id)
	if err != nil {
		return domain.Absence{}, err
	}
	if !domain.MayChange(kernel.PrincipalFrom(ctx), a.AccountID) {
		return domain.Absence{}, domain.ErrNotAllowed
	}
	return a, nil
}

func (s *Service) audit(ctx context.Context, action, id string) {
	if s.d.Audit != nil {
		s.d.Audit(ctx, action, id)
	}
}

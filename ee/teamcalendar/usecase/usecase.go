// Package usecase holds the team calendar use cases (#299): a person shares
// one of their own calendar subscriptions with a group or everyone, and the
// others see its events in their agenda.
//
// Ring: Use Cases.
package usecase

import (
	"context"

	"trisse-home/internal/shared/kernel"
)

// Subscription is an own calendar subscription and who else sees it.
type Subscription struct {
	ID, Name   string
	Visibility kernel.Visibility
}

// Sharing is the calendar's side (output port).
type Sharing interface {
	OwnSubscriptions(ctx context.Context) ([]Subscription, error)
	Share(ctx context.Context, id string, v kernel.Visibility) error
}

// Service implements the use cases.
type Service struct{ sharing Sharing }

// New returns the service.
func New(s Sharing) *Service { return &Service{sharing: s} }

// errNoPerson refuses guests, kiosk displays and tokens: they have no
// calendars of their own to share.
var errNoPerson = kernel.NewPublicError(kernel.ErrForbidden, "Only members and admins share calendars.")

func person(ctx context.Context) error {
	p := kernel.PrincipalFrom(ctx)
	if p.TabScoped() || p.IsToken() || p.AccountID == "" {
		return errNoPerson
	}
	return nil
}

// List returns the principal's own subscriptions and who sees them.
func (s *Service) List(ctx context.Context) ([]Subscription, error) {
	if err := person(ctx); err != nil {
		return nil, err
	}
	return s.sharing.OwnSubscriptions(ctx)
}

// Share sets who sees one of the principal's own subscriptions: private,
// everyone, or one of the principal's groups (admins: any group).
func (s *Service) Share(ctx context.Context, id, visibility string) error {
	if err := person(ctx); err != nil {
		return err
	}
	v, err := kernel.CheckShare(kernel.PrincipalFrom(ctx), visibility)
	if err != nil {
		return err
	}
	return s.sharing.Share(ctx, id, v)
}

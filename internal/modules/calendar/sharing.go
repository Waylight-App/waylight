package calendar

import (
	"context"
	"errors"

	"trisse-home/internal/modules/calendar/adapter/gateway"
	"trisse-home/internal/shared/kernel"
)

// Sharing connects the team calendar (ee/, #299) to the calendar without
// the core importing ee/, like the dashboard's team tabs: the composition
// root creates it before the modules, hands it to the calendar and, in
// builds with ee/, to the team calendar module, which switches calendars
// shared by others on while the licence allows it. Without that, a person's
// agenda holds only their own calendars.
type Sharing struct {
	enabled func(context.Context) bool
	repo    *gateway.SubscriptionRepository
}

// NewSharing returns a Sharing with shared calendars switched off.
func NewSharing() *Sharing { return &Sharing{} }

// Enable decides per request whether calendars shared with the principal
// are added to its agenda.
func (s *Sharing) Enable(on func(context.Context) bool) { s.enabled = on }

func (s *Sharing) on(ctx context.Context) bool { return s != nil && s.enabled != nil && s.enabled(ctx) }

// ErrNotReady is returned while the calendar module is switched off.
var ErrNotReady = errors.New("calendar not available")

// SubscriptionInfo is one of the principal's own subscriptions and who
// else sees it.
type SubscriptionInfo = gateway.SharedSubscription

// OwnSubscriptions lists the principal's own subscriptions with their
// visibility.
func (s *Sharing) OwnSubscriptions(ctx context.Context) ([]SubscriptionInfo, error) {
	if s == nil || s.repo == nil {
		return nil, ErrNotReady
	}
	return s.repo.Own(ctx)
}

// Share changes who sees one of the principal's own subscriptions; v is
// checked by the caller (kernel.CheckShare).
func (s *Sharing) Share(ctx context.Context, id string, v kernel.Visibility) error {
	if s == nil || s.repo == nil {
		return ErrNotReady
	}
	return s.repo.SetVisibility(ctx, id, v)
}

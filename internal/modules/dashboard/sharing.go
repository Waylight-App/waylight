package dashboard

import (
	"context"
	"errors"

	"waylight/internal/modules/dashboard/adapter/gateway"
	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
)

// Sharing connects team tabs (ee/, #295) to the dashboard without the core
// importing ee/. The composition root creates it before the modules, hands it
// to the dashboard and, in builds with ee/, to the team tabs module, which
// switches shared tabs on while the licence allows it.
type Sharing struct {
	enabled      func(context.Context) bool
	groupDefault func(context.Context) string
	repo         *gateway.Repository
}

// NewSharing returns a Sharing with shared tabs switched off.
func NewSharing() *Sharing { return &Sharing{} }

// Enable decides per request whether tabs shared with the principal are shown.
func (s *Sharing) Enable(on func(context.Context) bool) { s.enabled = on }

func (s *Sharing) on(ctx context.Context) bool { return s != nil && s.enabled != nil && s.enabled(ctx) }

// SetGroupDefault names, per request, the team tab that opens by default
// for the principal's groups when their own rules pick none.
func (s *Sharing) SetGroupDefault(fn func(context.Context) string) { s.groupDefault = fn }

func (s *Sharing) defaultTab(ctx context.Context) string {
	if !s.on(ctx) || s.groupDefault == nil {
		return ""
	}
	return s.groupDefault(ctx)
}

// TabInfo is an own tab and who else sees it.
type TabInfo struct {
	ID, Name   string
	Visibility kernel.Visibility
}

// ErrNotReady is returned before the dashboard module is built.
var ErrNotReady = errors.New("dashboard not ready")

// OwnTabs lists the principal's own tabs with their visibility.
func (s *Sharing) OwnTabs(ctx context.Context) ([]TabInfo, error) {
	if s.repo == nil {
		return nil, ErrNotReady
	}
	tabs, err := s.repo.OwnTabs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TabInfo, 0, len(tabs))
	for _, t := range tabs {
		out = append(out, TabInfo{ID: t.ID(), Name: t.Name(), Visibility: t.Access().Visibility})
	}
	return out, nil
}

// Share sets who besides the owner sees one of the principal's tabs:
// kernel.Private, kernel.Everyone or a group.
func (s *Sharing) Share(ctx context.Context, tabID string, v kernel.Visibility) error {
	if s.repo == nil {
		return ErrNotReady
	}
	return s.repo.SetTabVisibility(ctx, tabID, v)
}

// ErrTabNotFound is the dashboard's error for an unknown or foreign tab.
var ErrTabNotFound = domain.ErrTabNotFound

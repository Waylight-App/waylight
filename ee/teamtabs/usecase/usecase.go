// Package usecase holds the team tab use cases (#295): an admin shares one
// of their own dashboard tabs with a group or everyone, and members of that
// group see it read-only next to their own tabs.
//
// Ring: Use Cases.
package usecase

import (
	"context"

	"trisse-home/internal/shared/kernel"
)

// Tab is an own tab and who else sees it.
type Tab struct {
	ID, Name   string
	Visibility kernel.Visibility
}

// Sharing is the dashboard's side (output port): the admin's own tabs and
// setting their visibility.
type Sharing interface {
	OwnTabs(ctx context.Context) ([]Tab, error)
	Share(ctx context.Context, tabID string, v kernel.Visibility) error
}

// Service implements the team tab use cases.
type Service struct{ sharing Sharing }

// New returns the service.
func New(s Sharing) *Service { return &Service{sharing: s} }

// Tabs lists the admin's own tabs with their visibility.
func (s *Service) Tabs(ctx context.Context) ([]Tab, error) {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return nil, kernel.ErrAdminOnly
	}
	return s.sharing.OwnTabs(ctx)
}

// Share makes one of the admin's tabs a team tab for a group or everyone,
// or private again. Members of other groups and guests never see it.
func (s *Service) Share(ctx context.Context, tabID, visibility string) (kernel.Visibility, error) {
	p := kernel.PrincipalFrom(ctx)
	if !p.IsAdmin() {
		return "", kernel.ErrAdminOnly
	}
	v, err := kernel.CheckShare(p, visibility)
	if err != nil {
		return "", err
	}
	return v, s.sharing.Share(ctx, tabID, v)
}

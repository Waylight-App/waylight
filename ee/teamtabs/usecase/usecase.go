// Package usecase holds the team tab use cases (#295): an admin shares one
// of their own dashboard tabs with a group or everyone, and members of that
// group see it read-only next to their own tabs.
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"slices"

	"waylight/internal/shared/kernel"
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

// Defaults stores the default team tab per group (output port).
type Defaults interface {
	All(ctx context.Context) (map[string]string, error)
	// Set stores a tab for a group; "" removes the default.
	Set(ctx context.Context, groupID, tabID string) error
}

// Service implements the team tab use cases.
type Service struct {
	sharing  Sharing
	defaults Defaults
}

// New returns the service.
func New(s Sharing, d Defaults) *Service { return &Service{sharing: s, defaults: d} }

// ErrNotATeamTab is returned when a default names a tab that is not shared
// with the group (or everyone).
var ErrNotATeamTab = kernel.NewValidationError("tabId", "Choose one of your tabs that is shared with this group or with everyone.")

// Defaults lists the default team tab per group (admins).
func (s *Service) Defaults(ctx context.Context) (map[string]string, error) {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return nil, kernel.ErrAdminOnly
	}
	return s.defaults.All(ctx)
}

// SetDefault makes one of the admin's team tabs the default for a group;
// the tab must be shared with that group or everyone. "" removes it.
func (s *Service) SetDefault(ctx context.Context, groupID, tabID string) error {
	if !kernel.PrincipalFrom(ctx).IsAdmin() {
		return kernel.ErrAdminOnly
	}
	if tabID != "" {
		tabs, err := s.sharing.OwnTabs(ctx)
		if err != nil {
			return err
		}
		ok := false
		for _, t := range tabs {
			if t.ID == tabID && (t.Visibility == kernel.Everyone || t.Visibility == kernel.GroupVisibility(groupID)) {
				ok = true
			}
		}
		if !ok {
			return ErrNotATeamTab
		}
	}
	return s.defaults.Set(ctx, groupID, tabID)
}

// DefaultFor returns the default team tab for the principal: the one of
// the first of its groups, in group ID order, that has one. The dashboard
// still checks that the principal can see the tab.
func (s *Service) DefaultFor(ctx context.Context) string {
	groups := kernel.PrincipalFrom(ctx).Groups
	if len(groups) == 0 {
		return ""
	}
	all, err := s.defaults.All(ctx)
	if err != nil {
		return ""
	}
	sorted := slices.Clone(groups)
	slices.Sort(sorted)
	for _, g := range sorted {
		if id := all[g]; id != "" {
			return id
		}
	}
	return ""
}

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

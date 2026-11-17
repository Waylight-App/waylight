package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/shared/kernel"
)

// GroupView describes a group.
type GroupView struct {
	ID        string
	Name      string
	Accounts  []string
	CreatedAt time.Time
}

// GroupRequest is the input of Groups.Create and Groups.Update.
type GroupRequest struct {
	ID       string // empty on Create
	Name     string
	Accounts []string
}

// Groups manages groups (input port, admin only, #294). Records can be
// shared with a group once the modules store owners (ADR 0018).
type Groups interface {
	List(ctx context.Context) ([]GroupView, error)
	Create(ctx context.Context, req GroupRequest) (GroupView, error)
	Update(ctx context.Context, req GroupRequest) (GroupView, error)
	Delete(ctx context.Context, id string) error
	// Mine lists the groups the principal may share with: its own groups,
	// or every group for an admin (#301).
	Mine(ctx context.Context) ([]GroupView, error)
}

type groups struct{ deps Deps }

// NewGroups returns the group management use cases.
func NewGroups(deps Deps) Groups { return &groups{deps: deps} }

func groupView(g domain.Group) GroupView {
	return GroupView{ID: g.ID(), Name: g.Name(), Accounts: g.Accounts(), CreatedAt: g.CreatedAt()}
}

// List implements Groups.
func (uc *groups) List(ctx context.Context) ([]GroupView, error) {
	if err := adminOnly(ctx); err != nil {
		return nil, err
	}
	all, err := uc.deps.Groups.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	out := make([]GroupView, 0, len(all))
	for _, g := range all {
		out = append(out, groupView(g))
	}
	return out, nil
}

// checkAccounts refuses IDs that are not members or guests, so a group
// cannot name the owner, a kiosk display or an account that does not exist.
func (uc *groups) checkAccounts(ctx context.Context, ids []string) error {
	all, err := uc.deps.Accounts.List(ctx)
	if err != nil {
		return fmt.Errorf("list accounts: %w", err)
	}
	for _, id := range ids {
		if id == "" || id == domain.OwnerID {
			continue // dropped by Group.SetAccounts
		}
		ok := slices.ContainsFunc(all, func(a domain.Account) bool {
			return a.ID() == id && (a.IsPerson() || a.Role() == domain.RoleGuest)
		})
		if !ok {
			return domain.ErrMemberNotFound
		}
	}
	return nil
}

// Create implements Groups.
func (uc *groups) Create(ctx context.Context, req GroupRequest) (GroupView, error) {
	existing, err := uc.List(ctx)
	if err != nil {
		return GroupView{}, err
	}
	if len(existing) >= domain.MaxGroups {
		return GroupView{}, domain.ErrTooManyGroups
	}
	if err := uc.checkAccounts(ctx, req.Accounts); err != nil {
		return GroupView{}, err
	}
	g, err := domain.NewGroup(uc.deps.IDs.NewID(), req.Name, req.Accounts, uc.deps.Clock.Now())
	if err != nil {
		return GroupView{}, err
	}
	if err := uc.deps.Groups.Create(ctx, g); err != nil {
		return GroupView{}, err
	}
	uc.deps.Events.record(ctx, domain.EventGroupChanged, "", domain.Client{}, "created "+g.Name())
	return groupView(g), nil
}

// Update implements Groups.
func (uc *groups) Update(ctx context.Context, req GroupRequest) (GroupView, error) {
	if err := adminOnly(ctx); err != nil {
		return GroupView{}, err
	}
	g, err := uc.deps.Groups.Get(ctx, req.ID)
	if err != nil {
		return GroupView{}, err
	}
	if err := g.Rename(req.Name); err != nil {
		return GroupView{}, err
	}
	if err := uc.checkAccounts(ctx, req.Accounts); err != nil {
		return GroupView{}, err
	}
	g.SetAccounts(req.Accounts)
	if err := uc.deps.Groups.Update(ctx, g); err != nil {
		return GroupView{}, err
	}
	uc.deps.Events.record(ctx, domain.EventGroupChanged, "", domain.Client{}, "updated "+g.Name())
	return groupView(g), nil
}

// Delete implements Groups. Records shared with the group become visible to
// their owners only; the modules handle that once they store visibility.
func (uc *groups) Delete(ctx context.Context, id string) error {
	if err := adminOnly(ctx); err != nil {
		return err
	}
	g, err := uc.deps.Groups.Get(ctx, id)
	if errors.Is(err, domain.ErrGroupNotFound) {
		return err
	}
	if err != nil {
		return fmt.Errorf("load group: %w", err)
	}
	if err := uc.deps.Groups.Delete(ctx, id); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventGroupChanged, "", domain.Client{}, "deleted "+g.Name())
	return nil
}

// Mine implements Groups. It is self-service: members and guests learn the
// names of the groups they are in, nothing about other groups.
func (uc *groups) Mine(ctx context.Context) ([]GroupView, error) {
	if uc.deps.Groups == nil {
		return []GroupView{}, nil
	}
	p := kernel.PrincipalFrom(ctx)
	all, err := uc.deps.Groups.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	out := []GroupView{}
	for _, g := range all {
		if p.IsAdmin() || slices.Contains(p.Groups, g.ID()) {
			out = append(out, GroupView{ID: g.ID(), Name: g.Name(), Accounts: []string{}, CreatedAt: g.CreatedAt()})
		}
	}
	return out, nil
}

package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// MemberView describes a member or an admin for the admin's settings.
type MemberView struct {
	ID     string
	Name   string
	Groups []string
	// Role is "member" or "admin" (#434); RoleSource is "" when set by hand,
	// else the provider that granted the admin role ("sso", "directory").
	Role       string
	RoleSource string
	Disabled   bool
	CreatedAt  time.Time
}

// NewMemberRequest is the input of Members.Create.
type NewMemberRequest struct {
	Name     string
	Password string
}

// UpdateMemberRequest is the input of Members.Update; an empty password
// keeps it, an empty role too.
type UpdateMemberRequest struct {
	ID       string
	Disabled bool
	Password string
	// Role is "member" or "admin" (#434).
	Role string
}

// Members manages member accounts (input port, admin only, #294).
type Members interface {
	// Available reports whether members can be created in this version.
	Available() bool
	List(ctx context.Context) ([]MemberView, error)
	Create(ctx context.Context, req NewMemberRequest) (MemberView, error)
	Update(ctx context.Context, req UpdateMemberRequest) (MemberView, error)
	Delete(ctx context.Context, id string) error
}

type members struct {
	deps   Deps
	guests *guests // shares password hashing
}

// NewMembers returns the member management use cases.
func NewMembers(deps Deps) Members { return &members{deps: deps, guests: &guests{deps: deps}} }

// Available implements Members.
func (uc *members) Available() bool { return uc.deps.MembersEnabled }

func (uc *members) view(ctx context.Context, a domain.Account) (MemberView, error) {
	v := MemberView{ID: a.ID(), Name: a.Name(), Role: string(a.Role()), RoleSource: a.RoleSource(), Disabled: a.Disabled(),
		CreatedAt: a.CreatedAt(), Groups: []string{}}
	if uc.deps.Groups != nil {
		g, err := uc.deps.Groups.GroupsOf(ctx, a.ID())
		if err != nil {
			return MemberView{}, fmt.Errorf("load groups: %w", err)
		}
		v.Groups = g
	}
	return v, nil
}

// List implements Members.
func (uc *members) List(ctx context.Context) ([]MemberView, error) {
	if err := adminOnly(ctx); err != nil {
		return nil, err
	}
	all, err := uc.deps.Accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	out := []MemberView{}
	for _, a := range all {
		if !a.IsPerson() {
			continue
		}
		v, err := uc.view(ctx, a)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// Create implements Members.
func (uc *members) Create(ctx context.Context, req NewMemberRequest) (MemberView, error) {
	existing, err := uc.List(ctx)
	if err != nil {
		return MemberView{}, err
	}
	if !uc.deps.MembersEnabled {
		return MemberView{}, domain.ErrMembersUnavailable
	}
	if len(existing) >= domain.MaxMembers {
		return MemberView{}, domain.ErrTooManyMembers
	}
	hash, err := uc.guests.hash(ctx, req.Password)
	if err != nil {
		return MemberView{}, err
	}
	m, err := domain.NewMember(uc.deps.IDs.NewID(), req.Name, hash, uc.deps.Clock.Now())
	if err != nil {
		return MemberView{}, err
	}
	if err := uc.deps.Accounts.Create(ctx, m); err != nil {
		return MemberView{}, err
	}
	uc.deps.Events.record(ctx, domain.EventMemberChanged, m.ID(), domain.Client{}, "created "+m.Name())
	return uc.view(ctx, m)
}

func (uc *members) member(ctx context.Context, id string) (domain.Account, error) {
	if err := adminOnly(ctx); err != nil {
		return domain.Account{}, err
	}
	a, err := uc.deps.Accounts.Get(ctx, id)
	if errors.Is(err, domain.ErrGuestNotFound) || errors.Is(err, domain.ErrAccountNotFound) || (err == nil && !a.IsPerson()) {
		return domain.Account{}, domain.ErrMemberNotFound
	}
	return a, err
}

// Update implements Members. Disabling a member or changing its password
// ends its sessions.
func (uc *members) Update(ctx context.Context, req UpdateMemberRequest) (MemberView, error) {
	m, err := uc.member(ctx, req.ID)
	if err != nil {
		return MemberView{}, err
	}
	now := uc.deps.Clock.Now()
	self := kernel.PrincipalFrom(ctx).AccountID == m.ID()
	roleChanged := req.Role != "" && req.Role != string(m.Role())
	// An admin cannot take their own rights away or lock themselves out;
	// the owner or another admin does that (ADR 0022).
	if self && (req.Disabled || roleChanged) {
		return MemberView{}, domain.ErrNotOnYourself
	}
	if err := m.SetMemberAccess(req.Disabled, now); err != nil {
		return MemberView{}, err
	}
	if roleChanged {
		if err := m.SetRole(domain.Role(req.Role), domain.RoleSourceHand, now); err != nil {
			return MemberView{}, err
		}
	}
	if req.Password != "" {
		hash, err := uc.guests.hash(ctx, req.Password)
		if err != nil {
			return MemberView{}, err
		}
		if err := m.ChangePasswordHash(hash, now); err != nil {
			return MemberView{}, err
		}
	}
	if err := uc.deps.Accounts.Update(ctx, m); err != nil {
		return MemberView{}, fmt.Errorf("update member: %w", err)
	}
	// A demoted admin's sessions end too, so no open tab keeps admin pages.
	if req.Disabled || req.Password != "" || (roleChanged && req.Role == string(domain.RoleMember)) {
		if err := uc.deps.Sessions.DeleteAllFor(ctx, m.ID()); err != nil {
			return MemberView{}, fmt.Errorf("end member sessions: %w", err)
		}
	}
	detail := "updated " + m.Name()
	switch {
	case req.Disabled:
		detail = "disabled " + m.Name()
	case roleChanged && req.Role == string(domain.RoleAdmin):
		detail = "made " + m.Name() + " an admin"
	case roleChanged:
		detail = "made " + m.Name() + " a member"
	}
	uc.deps.Events.record(ctx, domain.EventMemberChanged, m.ID(), domain.Client{}, detail)
	return uc.view(ctx, m)
}

// Delete implements Members. The member's personal content is removed by
// the modules once they store owners (ADR 0018 follow-up work); until then
// members cannot be created, so there is none.
func (uc *members) Delete(ctx context.Context, id string) error {
	m, err := uc.member(ctx, id)
	if err != nil {
		return err
	}
	if kernel.PrincipalFrom(ctx).AccountID == m.ID() {
		return domain.ErrNotOnYourself
	}
	if err := uc.deps.Sessions.DeleteAllFor(ctx, m.ID()); err != nil {
		return fmt.Errorf("end member sessions: %w", err)
	}
	if uc.deps.Passkeys != nil {
		if err := uc.deps.Passkeys.DeleteByAccount(ctx, m.ID()); err != nil {
			return err
		}
	}
	if uc.deps.Groups != nil {
		if err := uc.deps.Groups.RemoveAccount(ctx, m.ID()); err != nil {
			return fmt.Errorf("leave groups: %w", err)
		}
	}
	if err := uc.deps.Accounts.Delete(ctx, m.ID()); err != nil {
		return err
	}
	uc.deps.Hooks.accountDeleted(ctx, m.ID())
	uc.deps.Events.record(ctx, domain.EventMemberChanged, m.ID(), domain.Client{}, "deleted "+m.Name())
	return nil
}

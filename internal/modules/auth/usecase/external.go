package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"trisse-home/internal/modules/auth/domain"
)

// Provisioning manages member accounts on behalf of an identity provider
// (input port, #302). The ee/ single sign-on uses it to create a member on
// first sign-in, keep their groups in step with the provider and lock them
// out when the provider no longer lets them in. The rules of the auth module
// still apply: members must be switched on, MaxMembers holds, and only
// member accounts can be changed this way.
type Provisioning interface {
	// CreateMember adds a member that can only sign in through the provider:
	// its password is random and never shown. A taken name gets a number.
	CreateMember(ctx context.Context, name string) (accountID string, err error)
	// UpdateGroups adds the member to the named groups (creating missing
	// ones) and removes them from the named groups to leave.
	UpdateGroups(ctx context.Context, accountID string, join, leave []string) error
	// SetDisabled locks a member out or lets them in again.
	SetDisabled(ctx context.Context, accountID string, disabled bool) error
	// EndSessions signs the member out everywhere.
	EndSessions(ctx context.Context, accountID string) error

	// The operations below serve directory provisioning (SCIM, #396), where
	// the directory owns members and groups by ID.

	// Member returns a member; domain.ErrMemberNotFound for anything else.
	Member(ctx context.Context, accountID string) (ProvisionedMember, error)
	// Members lists all members.
	Members(ctx context.Context) ([]ProvisionedMember, error)
	// DeleteMember removes a member as the settings do: sessions end, the
	// member leaves its groups, the account is deleted.
	DeleteMember(ctx context.Context, accountID string) error
	// Groups lists all groups.
	Groups(ctx context.Context) ([]domain.Group, error)
	// SaveGroup creates (empty id) or replaces a group's name and members;
	// only members can be in a group set this way.
	SaveGroup(ctx context.Context, id, name string, accounts []string) (domain.Group, error)
	// DeleteGroup removes a group.
	DeleteGroup(ctx context.Context, id string) error
}

// ProvisionedMember is a member as a directory sees it.
type ProvisionedMember struct {
	ID       string
	Name     string
	Disabled bool
	Groups   []string // group IDs
}

type provisioning struct{ deps Deps }

// NewProvisioning returns the use cases.
func NewProvisioning(deps Deps) Provisioning { return provisioning{deps: deps} }

// errNotMember refuses changes to the owner or guests through the provider.
var errNotMember = errors.New("only member accounts are managed by the identity provider")

func (uc provisioning) member(ctx context.Context, id string) (domain.Account, error) {
	a, err := uc.deps.Accounts.Get(ctx, id)
	if err != nil {
		return domain.Account{}, err
	}
	if a.Role() != domain.RoleMember {
		return domain.Account{}, errNotMember
	}
	return a, nil
}

// CreateMember implements Provisioning.
func (uc provisioning) CreateMember(ctx context.Context, name string) (string, error) {
	if !uc.deps.MembersEnabled {
		return "", domain.ErrMembersUnavailable
	}
	all, err := uc.deps.Accounts.List(ctx)
	if err != nil {
		return "", fmt.Errorf("list accounts: %w", err)
	}
	members, names := 0, map[string]bool{}
	for _, a := range all {
		names[strings.ToLower(a.Name())] = true
		if a.Role() == domain.RoleMember {
			members++
		}
	}
	if members >= domain.MaxMembers {
		return "", domain.ErrTooManyMembers
	}
	// A password nobody knows: the account signs in through the provider only.
	secret, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	hash, err := uc.deps.Hasher.Hash(secret)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	base := accountName(name)
	candidate := base
	for i := 2; names[candidate]; i++ {
		suffix := "-" + strconv.Itoa(i)
		candidate = base[:min(len(base), maxNameLen-len(suffix))] + suffix
	}
	m, err := domain.NewMember(uc.deps.IDs.NewID(), candidate, hash, uc.deps.Clock.Now())
	if err != nil {
		return "", err
	}
	if err := uc.deps.Accounts.Create(ctx, m); err != nil {
		return "", err
	}
	uc.deps.Events.record(ctx, domain.EventMemberChanged, m.ID(), domain.Client{}, "created "+m.Name()+" through single sign-on")
	return m.ID(), nil
}

// maxNameLen is the longest account name the domain accepts.
const maxNameLen = 32

// accountName turns what the provider calls a person ("Anna Weber",
// "anna.weber@example.org") into a valid account name: lowercase letters,
// digits, dots, dashes and underscores, 2 to 32 characters.
func accountName(raw string) string {
	raw, _, _ = strings.Cut(strings.ToLower(strings.TrimSpace(raw)), "@")
	var b strings.Builder
	dash := false
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	name := strings.Trim(b.String(), "-.")
	if len(name) > maxNameLen {
		name = strings.Trim(name[:maxNameLen], "-.")
	}
	if len(name) < 2 {
		return "member"
	}
	return name
}

// UpdateGroups implements Provisioning.
func (uc provisioning) UpdateGroups(ctx context.Context, accountID string, join, leave []string) error {
	if uc.deps.Groups == nil {
		return nil
	}
	if _, err := uc.member(ctx, accountID); err != nil {
		return err
	}
	groups, err := uc.deps.Groups.List(ctx)
	if err != nil {
		return fmt.Errorf("list groups: %w", err)
	}
	byName := map[string]domain.Group{}
	for _, g := range groups {
		byName[strings.ToLower(g.Name())] = g
	}
	for _, name := range join {
		g, ok := byName[strings.ToLower(name)]
		if !ok {
			created, err := domain.NewGroup(uc.deps.IDs.NewID(), name, []string{accountID}, uc.deps.Clock.Now())
			if err != nil {
				return err
			}
			if err := uc.deps.Groups.Create(ctx, created); err != nil {
				return err
			}
			byName[strings.ToLower(name)] = created
			continue
		}
		if !slices.Contains(g.Accounts(), accountID) {
			g.SetAccounts(append(g.Accounts(), accountID))
			if err := uc.deps.Groups.Update(ctx, g); err != nil {
				return err
			}
		}
	}
	for _, name := range leave {
		g, ok := byName[strings.ToLower(name)]
		if !ok || !slices.Contains(g.Accounts(), accountID) {
			continue
		}
		g.SetAccounts(slices.DeleteFunc(g.Accounts(), func(id string) bool { return id == accountID }))
		if err := uc.deps.Groups.Update(ctx, g); err != nil {
			return err
		}
	}
	return nil
}

// SetDisabled implements Provisioning.
func (uc provisioning) SetDisabled(ctx context.Context, accountID string, disabled bool) error {
	m, err := uc.member(ctx, accountID)
	if err != nil {
		return err
	}
	if m.Disabled() == disabled {
		return nil
	}
	if err := m.SetMemberAccess(disabled, uc.deps.Clock.Now()); err != nil {
		return err
	}
	if err := uc.deps.Accounts.Update(ctx, m); err != nil {
		return fmt.Errorf("update account: %w", err)
	}
	if disabled {
		return uc.EndSessions(ctx, accountID)
	}
	return nil
}

// EndSessions implements Provisioning.
func (uc provisioning) EndSessions(ctx context.Context, accountID string) error {
	if _, err := uc.member(ctx, accountID); err != nil {
		return err
	}
	if err := uc.deps.Sessions.DeleteAllFor(ctx, accountID); err != nil {
		return fmt.Errorf("end sessions: %w", err)
	}
	return nil
}

// Member implements Provisioning.
func (uc provisioning) Member(ctx context.Context, accountID string) (ProvisionedMember, error) {
	a, err := uc.member(ctx, accountID)
	if err != nil {
		return ProvisionedMember{}, domain.ErrMemberNotFound
	}
	return uc.view(ctx, a)
}

func (uc provisioning) view(ctx context.Context, a domain.Account) (ProvisionedMember, error) {
	m := ProvisionedMember{ID: a.ID(), Name: a.Name(), Disabled: a.Disabled(), Groups: []string{}}
	if uc.deps.Groups != nil {
		ids, err := uc.deps.Groups.GroupsOf(ctx, a.ID())
		if err != nil {
			return ProvisionedMember{}, fmt.Errorf("groups of member: %w", err)
		}
		m.Groups = append(m.Groups, ids...)
	}
	return m, nil
}

// Members implements Provisioning.
func (uc provisioning) Members(ctx context.Context) ([]ProvisionedMember, error) {
	all, err := uc.deps.Accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	out := []ProvisionedMember{}
	for _, a := range all {
		if a.Role() != domain.RoleMember {
			continue
		}
		m, err := uc.view(ctx, a)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

// DeleteMember implements Provisioning.
func (uc provisioning) DeleteMember(ctx context.Context, accountID string) error {
	m, err := uc.member(ctx, accountID)
	if err != nil {
		return domain.ErrMemberNotFound
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
	uc.deps.Events.record(ctx, domain.EventMemberChanged, m.ID(), domain.Client{}, "deleted "+m.Name()+" through directory provisioning")
	return nil
}

// Groups implements Provisioning.
func (uc provisioning) Groups(ctx context.Context) ([]domain.Group, error) {
	if uc.deps.Groups == nil {
		return []domain.Group{}, nil
	}
	return uc.deps.Groups.List(ctx)
}

// SaveGroup implements Provisioning.
func (uc provisioning) SaveGroup(ctx context.Context, id, name string, accounts []string) (domain.Group, error) {
	if uc.deps.Groups == nil {
		return domain.Group{}, domain.ErrMembersUnavailable
	}
	for _, a := range accounts {
		if _, err := uc.member(ctx, a); err != nil {
			return domain.Group{}, domain.ErrMemberNotFound
		}
	}
	if id == "" {
		g, err := domain.NewGroup(uc.deps.IDs.NewID(), name, accounts, uc.deps.Clock.Now())
		if err != nil {
			return domain.Group{}, err
		}
		if err := uc.deps.Groups.Create(ctx, g); err != nil {
			return domain.Group{}, err
		}
		return g, nil
	}
	g, err := uc.deps.Groups.Get(ctx, id)
	if err != nil {
		return domain.Group{}, err
	}
	if name != g.Name() {
		if err := g.Rename(name); err != nil {
			return domain.Group{}, err
		}
	}
	g.SetAccounts(accounts)
	if err := uc.deps.Groups.Update(ctx, g); err != nil {
		return domain.Group{}, err
	}
	return g, nil
}

// DeleteGroup implements Provisioning.
func (uc provisioning) DeleteGroup(ctx context.Context, id string) error {
	if uc.deps.Groups == nil {
		return domain.ErrMembersUnavailable
	}
	if _, err := uc.deps.Groups.Get(ctx, id); err != nil {
		return err
	}
	return uc.deps.Groups.Delete(ctx, id)
}

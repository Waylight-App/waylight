package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/shared/kernel"
)

// memGroups is an in-memory GroupRepository.
type memGroups struct{ groups map[string]domain.Group }

func newMemGroups() *memGroups { return &memGroups{groups: map[string]domain.Group{}} }

func (m *memGroups) List(context.Context) ([]domain.Group, error) {
	var out []domain.Group
	for _, g := range m.groups {
		out = append(out, g)
	}
	return out, nil
}
func (m *memGroups) Get(_ context.Context, id string) (domain.Group, error) {
	g, ok := m.groups[id]
	if !ok {
		return domain.Group{}, domain.ErrGroupNotFound
	}
	return g, nil
}
func (m *memGroups) Create(_ context.Context, g domain.Group) error {
	for _, o := range m.groups {
		if o.Name() == g.Name() {
			return domain.ErrGroupNameTaken
		}
	}
	m.groups[g.ID()] = g
	return nil
}
func (m *memGroups) Update(_ context.Context, g domain.Group) error { m.groups[g.ID()] = g; return nil }
func (m *memGroups) Delete(_ context.Context, id string) error {
	delete(m.groups, id)
	return nil
}
func (m *memGroups) GroupsOf(_ context.Context, account string) ([]string, error) {
	out := []string{}
	for id, g := range m.groups {
		if slices.Contains(g.Accounts(), account) {
			out = append(out, id)
		}
	}
	slices.Sort(out)
	return out, nil
}
func (m *memGroups) RemoveAccount(_ context.Context, account string) error {
	for id, g := range m.groups {
		g.SetAccounts(slices.DeleteFunc(g.Accounts(), func(a string) bool { return a == account }))
		m.groups[id] = g
	}
	return nil
}

func TestMembersAreOffUntilEnabled(t *testing.T) {
	e := newEnv().withAccount()
	members := NewMembers(e.deps)
	if members.Available() {
		t.Fatal("members must be off by default")
	}
	if _, err := members.Create(ctx, NewMemberRequest{Name: "anna", Password: "member-password-1"}); !errors.Is(err, domain.ErrMembersUnavailable) {
		t.Fatalf("created a member while off: %v", err)
	}
}

func TestMemberLifecycleAndGroups(t *testing.T) {
	e := newEnv().withAccount()
	e.deps.MembersEnabled = true
	groupsRepo := newMemGroups()
	e.deps.Groups = groupsRepo
	members, groups := NewMembers(e.deps), NewGroups(e.deps)

	m, err := members.Create(ctx, NewMemberRequest{Name: "anna", Password: "member-password-1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewGuests(e.deps).Create(ctx, NewGuestRequest{Name: "anna", Password: "guest-password-1"}); !errors.Is(err, domain.ErrGuestNameTaken) {
		t.Fatalf("guests and members share names: %v", err)
	}
	office, err := groups.Create(ctx, GroupRequest{Name: "Office", Accounts: []string{m.ID}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := groups.Create(ctx, GroupRequest{Name: "Bad", Accounts: []string{domain.OwnerID}}); err != nil {
		t.Fatalf("the owner is dropped from groups, not an error: %v", err)
	}
	if _, err := groups.Create(ctx, GroupRequest{Name: "Ghost", Accounts: []string{"nobody"}}); !errors.Is(err, domain.ErrMemberNotFound) {
		t.Fatalf("unknown account in group: %v", err)
	}

	// The member signs in with a name and carries its groups.
	session, err := NewLogin(e.deps).Execute(ctx, LoginRequest{Username: "anna", Password: "member-password-1", ClientKey: "a"})
	if err != nil {
		t.Fatal(err)
	}
	auth := NewAuthenticate(e.sessions, e.accounts, groupsRepo, nil, e.clock, e.deps.TTL)
	res, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: session.SessionToken})
	if err != nil || res.Principal.Role != "member" || res.Principal.AccountID != m.ID || !slices.Equal(res.Principal.Groups, []string{office.ID}) {
		t.Fatalf("principal %+v %v", res.Principal, err)
	}

	// A member manages nothing of the instance.
	memberCtx := kernel.WithPrincipal(ctx, res.Principal)
	if _, err := members.List(memberCtx); !errors.Is(err, domain.ErrOwnerOnly) {
		t.Fatalf("member listed members: %v", err)
	}
	if _, err := groups.List(memberCtx); !errors.Is(err, domain.ErrOwnerOnly) {
		t.Fatalf("member listed groups: %v", err)
	}
	if _, err := NewGuests(e.deps).List(memberCtx); !errors.Is(err, domain.ErrOwnerOnly) {
		t.Fatalf("member listed guests: %v", err)
	}

	// Disabling ends the session.
	if _, err := members.Update(ctx, UpdateMemberRequest{ID: m.ID, Disabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Execute(ctx, AuthenticateRequest{SessionToken: session.SessionToken}); !errors.Is(err, domain.ErrSessionInvalid) {
		t.Fatalf("disabled member still signed in: %v", err)
	}
	if _, err := members.Update(ctx, UpdateMemberRequest{ID: domain.OwnerID}); !errors.Is(err, domain.ErrMemberNotFound) {
		t.Fatalf("owner updated as member: %v", err)
	}

	// Deleting takes the member out of its groups.
	if err := members.Delete(ctx, m.ID); err != nil {
		t.Fatal(err)
	}
	if g, _ := groupsRepo.Get(ctx, office.ID); len(g.Accounts()) != 0 {
		t.Fatalf("deleted member still in group: %v", g.Accounts())
	}
	if list, _ := members.List(ctx); len(list) != 0 {
		t.Fatalf("left: %v", list)
	}
	if err := groups.Delete(ctx, office.ID); err != nil {
		t.Fatal(err)
	}
}

type kindScope map[string]string

func (kindScope) ReachForTabs(context.Context, []string) (domain.GuestReach, error) {
	return domain.GuestReach{}, nil
}
func (k kindScope) DataKinds() map[string]string { return k }

func TestAuthorizeMember(t *testing.T) {
	a := NewAuthorizer(kindScope{"notes": domain.KindPersonalContent, "services": "instance"})
	member := kernel.Principal{AccountID: "m1", Role: "member"}
	if err := a.Authorize(ctx, member, "POST", "/api/notes"); err != nil {
		t.Fatal(err)
	}
	for _, c := range [][2]string{{"POST", "/api/services"}, {"GET", "/api/auth/guests"}, {"POST", "/api/auth/members"}} {
		if err := a.Authorize(ctx, member, c[0], c[1]); !errors.Is(err, domain.ErrMemberNotAllowed) {
			t.Errorf("%v: %v", c, err)
		}
	}
	if err := a.Authorize(ctx, kernel.Principal{AccountID: "x", Role: "somethingnew"}, "GET", "/api/notes"); err == nil {
		t.Fatal("an unknown role must be refused")
	}
	if err := a.Authorize(ctx, kernel.Principal{AccountID: "owner", Role: "owner"}, "DELETE", "/api/anything"); err != nil {
		t.Fatal(err)
	}
}

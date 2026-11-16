package usecase

import (
	"errors"
	"slices"
	"testing"

	"trisse-home/internal/modules/auth/domain"
)

func provisioningEnv(t *testing.T) (*env, *memGroups, Provisioning) {
	t.Helper()
	e := newEnv().withAccount()
	groups := newMemGroups()
	e.deps.Groups, e.deps.MembersEnabled = groups, true
	return e, groups, NewProvisioning(e.deps)
}

func groupNamesOf(g *memGroups, account string) []string {
	var out []string
	for _, gr := range g.groups {
		if slices.Contains(gr.Accounts(), account) {
			out = append(out, gr.Name())
		}
	}
	slices.Sort(out)
	return out
}

func TestProvisioningCreatesMembersWithFreeNames(t *testing.T) {
	e, _, p := provisioningEnv(t)
	first, err := p.CreateMember(ctx, "anna")
	if err != nil {
		t.Fatal(err)
	}
	second, err := p.CreateMember(ctx, "anna@example.org")
	if err != nil {
		t.Fatal(err)
	}
	a, _ := e.accounts.Get(ctx, first)
	b, _ := e.accounts.Get(ctx, second)
	if a.Name() != "anna" || b.Name() != "anna-2" || a.Role() != domain.RoleMember {
		t.Fatalf("names %q, %q", a.Name(), b.Name())
	}
}

func TestAccountNameFromProviderNames(t *testing.T) {
	for raw, want := range map[string]string{
		"Anna Weber":             "anna-weber",
		"anna.weber@example.org": "anna.weber",
		"Jörg Ünal":              "j-rg-nal",
		"!":                      "member",
		"a-very-long-name-that-does-not-fit-at-all": "a-very-long-name-that-does-not-f",
	} {
		if got := accountName(raw); got != want {
			t.Errorf("%q: got %q, want %q", raw, got, want)
		}
	}
}

func TestProvisioningNeedsMembersSwitchedOn(t *testing.T) {
	e, _, _ := provisioningEnv(t)
	e.deps.MembersEnabled = false
	if _, err := NewProvisioning(e.deps).CreateMember(ctx, "anna"); !errors.Is(err, domain.ErrMembersUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestProvisioningKeepsGroupsInStep(t *testing.T) {
	_, groups, p := provisioningEnv(t)
	id, _ := p.CreateMember(ctx, "anna")
	if err := p.UpdateGroups(ctx, id, []string{"Office", "Sales"}, nil); err != nil {
		t.Fatal(err)
	}
	if got := groupNamesOf(groups, id); !slices.Equal(got, []string{"Office", "Sales"}) {
		t.Fatalf("after join: %v", got)
	}
	// Names match without regard to case, so the provider's spelling does not duplicate groups.
	if err := p.UpdateGroups(ctx, id, []string{"office"}, []string{"Sales"}); err != nil {
		t.Fatal(err)
	}
	if got := groupNamesOf(groups, id); !slices.Equal(got, []string{"Office"}) {
		t.Fatalf("after leave: %v", got)
	}
	if len(groups.groups) != 2 {
		t.Fatalf("groups: %d", len(groups.groups))
	}
}

func TestProvisioningLocksMembersOutAndEndsSessions(t *testing.T) {
	e, _, p := provisioningEnv(t)
	id, _ := p.CreateMember(ctx, "anna")
	if _, err := e.deps.issuer().issue(ctx, id, domain.Client{IP: "10.0.0.9"}); err != nil {
		t.Fatal(err)
	}
	if err := p.SetDisabled(ctx, id, true); err != nil {
		t.Fatal(err)
	}
	if a, _ := e.accounts.Get(ctx, id); !a.Disabled() {
		t.Fatal("not disabled")
	}
	for _, s := range e.sessions.byHash {
		if s.AccountID() == id {
			t.Fatal("session survived")
		}
	}
}

func TestProvisioningLeavesTheOwnerAlone(t *testing.T) {
	_, _, p := provisioningEnv(t)
	if err := p.SetDisabled(ctx, domain.OwnerID, true); err == nil {
		t.Fatal("owner disabled through the provider")
	}
	if err := p.UpdateGroups(ctx, domain.OwnerID, []string{"Office"}, nil); err == nil {
		t.Fatal("owner's groups changed through the provider")
	}
}

func TestProvisioningForDirectories(t *testing.T) {
	e, groups, p := provisioningEnv(t)
	anna, _ := p.CreateMember(ctx, "anna")
	ben, _ := p.CreateMember(ctx, "ben")

	g, err := p.SaveGroup(ctx, "", "Office", []string{anna})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.SaveGroup(ctx, g.ID(), "Front office", []string{anna, ben}); err != nil {
		t.Fatal(err)
	}
	m, err := p.Member(ctx, ben)
	if err != nil || m.Name != "ben" || !slices.Equal(m.Groups, []string{g.ID()}) {
		t.Fatalf("member: %+v %v", m, err)
	}
	if list, _ := p.Members(ctx); len(list) != 2 {
		t.Fatalf("members: %v", list)
	}
	// Only members can be put into a group from a directory.
	if _, err := p.SaveGroup(ctx, g.ID(), "Front office", []string{domain.OwnerID}); !errors.Is(err, domain.ErrMemberNotFound) {
		t.Fatalf("owner in a directory group: %v", err)
	}
	if err := p.DeleteMember(ctx, ben); err != nil {
		t.Fatal(err)
	}
	if _, err := e.accounts.Get(ctx, ben); err == nil {
		t.Fatal("member still there")
	}
	if got := groupNamesOf(groups, ben); len(got) != 0 {
		t.Fatalf("deleted member still in %v", got)
	}
	if err := p.DeleteGroup(ctx, g.ID()); err != nil || len(groups.groups) != 0 {
		t.Fatalf("delete group: %v, %d left", err, len(groups.groups))
	}
	if _, err := p.Member(ctx, domain.OwnerID); !errors.Is(err, domain.ErrMemberNotFound) {
		t.Fatalf("owner as member: %v", err)
	}
}

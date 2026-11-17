package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	licence "waylight/ee/licence/domain"
	"waylight/ee/sso/domain"
	authdomain "waylight/internal/modules/auth/domain"
)

type memLinks map[string]Link

func (m memLinks) Get(_ context.Context, s string) (Link, bool, error) {
	l, ok := m[s]
	return l, ok, nil
}
func (m memLinks) Save(_ context.Context, l Link) error     { m[l.Subject] = l; return nil }
func (m memLinks) Delete(_ context.Context, s string) error { delete(m, s); return nil }

type member struct {
	groups   []string
	disabled bool
	// admin is set through SetAdmin; adminSource remembers who granted it.
	admin       bool
	adminSource string
}

// SetAdmin mirrors the auth module: a source takes back only what it gave.
func (f *fakeAccounts) SetAdmin(_ context.Context, id string, admin bool, source string) (bool, error) {
	m, ok := f.members[id]
	if !ok {
		return false, authdomain.ErrGuestNotFound
	}
	switch {
	case admin && !m.admin:
		m.admin, m.adminSource = true, source
	case !admin && m.admin && m.adminSource == source:
		m.admin, m.adminSource = false, ""
	default:
		return false, nil
	}
	return true, nil
}

type fakeAccounts struct {
	members map[string]*member
	next    int
}

func (f *fakeAccounts) CreateMember(_ context.Context, _ string) (string, error) {
	f.next++
	id := "m" + string(rune('0'+f.next))
	f.members[id] = &member{}
	return id, nil
}

func (f *fakeAccounts) get(id string) (*member, error) {
	m, ok := f.members[id]
	if !ok {
		return nil, authdomain.ErrGuestNotFound
	}
	return m, nil
}

func (f *fakeAccounts) UpdateGroups(_ context.Context, id string, join, leave []string) error {
	m, err := f.get(id)
	if err != nil {
		return err
	}
	for _, g := range join {
		if !slices.Contains(m.groups, g) {
			m.groups = append(m.groups, g)
		}
	}
	m.groups = slices.DeleteFunc(m.groups, func(g string) bool { return slices.Contains(leave, g) })
	return nil
}

func (f *fakeAccounts) SetDisabled(_ context.Context, id string, d bool) error {
	m, err := f.get(id)
	if err != nil {
		return err
	}
	m.disabled = d
	return nil
}

type level licence.Level

func (l *level) Level(context.Context, string) licence.Level { return licence.Level(*l) }

var ctx = context.Background()

func setup() (*Resolver, *fakeAccounts, memLinks, *level) {
	accounts, links, lvl := &fakeAccounts{members: map[string]*member{}}, memLinks{}, level(licence.LevelFull)
	return New(Deps{Links: links, Accounts: accounts, Licensing: &lvl,
		Policy: domain.Policy{Emails: []string{"@example.org"}, Groups: []string{"waylight-users"}}}), accounts, links, &lvl
}

func anna(groups ...string) domain.Identity {
	return domain.Identity{Subject: "anna-sub", Email: "anna@example.org", EmailVerified: true, Username: "anna", Groups: groups}
}

func TestFirstSignInCreatesTheMemberWithGroups(t *testing.T) {
	r, accounts, _, _ := setup()
	id, err := r.Resolve(ctx, anna("Office"))
	if err != nil {
		t.Fatal(err)
	}
	if again, _ := r.Resolve(ctx, anna("Office")); again != id {
		t.Fatalf("second sign-in got %q, want %q", again, id)
	}
	if len(accounts.members) != 1 || !slices.Equal(accounts.members[id].groups, []string{"Office"}) {
		t.Fatalf("members %+v", accounts.members)
	}
}

func TestGroupsFollowTheProviderButHandAddedOnesStay(t *testing.T) {
	r, accounts, _, _ := setup()
	id, _ := r.Resolve(ctx, anna("Office", "Sales"))
	accounts.members[id].groups = append(accounts.members[id].groups, "Added-by-admin")
	if _, err := r.Resolve(ctx, anna("Office")); err != nil {
		t.Fatal(err)
	}
	if got := accounts.members[id].groups; !slices.Equal(got, []string{"Office", "Added-by-admin"}) {
		t.Fatalf("groups %v", got)
	}
}

func TestLosingAccessLocksOutAndComingBackLetsIn(t *testing.T) {
	r, accounts, _, _ := setup()
	id, _ := r.Resolve(ctx, anna())
	gone := domain.Identity{Subject: "anna-sub", Email: "anna@elsewhere.test", EmailVerified: true}
	if _, err := r.Resolve(ctx, gone); !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("after losing access: %v", err)
	}
	if !accounts.members[id].disabled {
		t.Fatal("not locked out")
	}
	if got, err := r.Resolve(ctx, anna()); err != nil || got != id || accounts.members[id].disabled {
		t.Fatalf("back again: %q %v disabled=%v", got, err, accounts.members[id].disabled)
	}
}

func TestAnAdminsLockOutIsNotUndone(t *testing.T) {
	r, accounts, links, _ := setup()
	id, _ := r.Resolve(ctx, anna())
	accounts.members[id].disabled = true // an admin locked the member out
	if _, err := r.Resolve(ctx, anna()); err != nil {
		t.Fatal(err)
	}
	if !accounts.members[id].disabled || links["anna-sub"].DisabledBySSO {
		t.Fatal("single sign-on undid the admin's decision")
	}
}

func TestWithoutLicenceOnlyTheOwnerSignsIn(t *testing.T) {
	r, accounts, _, lvl := setup()
	*lvl = level(licence.LevelNone)
	if _, err := r.Resolve(ctx, anna()); !errors.Is(err, ErrNotLicensed) {
		t.Fatalf("got %v", err)
	}
	if len(accounts.members) != 0 {
		t.Fatal("member created without licence")
	}
}

func TestReadOnlyLicenceLetsExistingMembersInButCreatesNobody(t *testing.T) {
	r, accounts, _, lvl := setup()
	id, _ := r.Resolve(ctx, anna("Office"))
	*lvl = level(licence.LevelRead)
	if got, err := r.Resolve(ctx, anna("Sales")); err != nil || got != id {
		t.Fatalf("existing member: %q %v", got, err)
	}
	if got := accounts.members[id].groups; !slices.Equal(got, []string{"Office"}) {
		t.Fatalf("groups changed on a read-only licence: %v", got)
	}
	bea := domain.Identity{Subject: "bea-sub", Email: "bea@example.org", EmailVerified: true}
	if _, err := r.Resolve(ctx, bea); !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("new member on read-only licence: %v", err)
	}
}

func TestADeletedMemberIsCreatedAgain(t *testing.T) {
	r, accounts, _, _ := setup()
	id, _ := r.Resolve(ctx, anna())
	delete(accounts.members, id)
	again, err := r.Resolve(ctx, anna())
	if err != nil || again == id {
		t.Fatalf("got %q %v", again, err)
	}
}

func TestAdminGroupGrantsAndTakesBackOnlyItsOwnRole(t *testing.T) {
	accounts, lvl := &fakeAccounts{members: map[string]*member{}}, level(licence.LevelFull)
	r := New(Deps{Links: memLinks{}, Accounts: accounts, Licensing: &lvl,
		Policy: domain.Policy{Emails: []string{"@example.org"}, AdminGroups: []string{"waylight-admins"}}})
	id, err := r.Resolve(ctx, anna("waylight-admins"))
	if err != nil || !accounts.members[id].admin || accounts.members[id].adminSource != "sso" {
		t.Fatalf("first sign-in in the admin group: %v %+v", err, accounts.members[id])
	}
	if _, err := r.Resolve(ctx, anna("Office")); err != nil || accounts.members[id].admin {
		t.Fatalf("after leaving the group: %v %+v", err, accounts.members[id])
	}
	// An admin made by hand keeps the role whatever the provider says.
	accounts.members[id].admin, accounts.members[id].adminSource = true, ""
	if _, err := r.Resolve(ctx, anna()); err != nil || !accounts.members[id].admin {
		t.Fatalf("hand-made admin demoted: %v %+v", err, accounts.members[id])
	}
}

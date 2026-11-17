package usecase

import (
	"context"
	"errors"
	"slices"
	"testing"

	"waylight/ee/ldap/domain"
	licence "waylight/ee/licence/domain"
	authdomain "waylight/internal/modules/auth/domain"
)

type fakeDir struct {
	users    map[string]domain.User // name -> user
	password map[string]string      // DN -> password
	down     bool
}

func (f *fakeDir) Lookup(_ context.Context, name string) (domain.User, bool, error) {
	if f.down {
		return domain.User{}, false, errors.New("unreachable")
	}
	u, ok := f.users[name]
	return u, ok, nil
}

func (f *fakeDir) Verify(_ context.Context, dn, pw string) error {
	if f.password[dn] != pw || pw == "" {
		return errors.New("invalid credentials")
	}
	return nil
}

type memLinks map[string]Link

func (m memLinks) Get(_ context.Context, n string) (Link, bool, error) {
	l, ok := m[n]
	return l, ok, nil
}
func (m memLinks) Save(_ context.Context, l Link) error     { m[l.Name] = l; return nil }
func (m memLinks) Delete(_ context.Context, n string) error { delete(m, n); return nil }

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

type fakeAccounts struct{ members map[string]*member }

func (f *fakeAccounts) CreateMember(_ context.Context, name string) (string, error) {
	id := "acc-" + name
	f.members[id] = &member{}
	return id, nil
}

func (f *fakeAccounts) UpdateGroups(_ context.Context, id string, join, leave []string) error {
	m, ok := f.members[id]
	if !ok {
		return authdomain.ErrGuestNotFound
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
	m, ok := f.members[id]
	if !ok {
		return authdomain.ErrGuestNotFound
	}
	m.disabled = d
	return nil
}

type level licence.Level

func (l *level) Level(context.Context, string) licence.Level { return licence.Level(*l) }

var ctx = context.Background()

const annaDN = "uid=anna,ou=people,dc=example,dc=org"

func setup() (*Authenticator, *fakeDir, *fakeAccounts, memLinks, *level) {
	dir := &fakeDir{
		users:    map[string]domain.User{"anna": {DN: annaDN, Groups: []string{"cn=Office,ou=groups,dc=example,dc=org"}}},
		password: map[string]string{annaDN: "anna-pw"},
	}
	accounts, links, lvl := &fakeAccounts{members: map[string]*member{}}, memLinks{}, level(licence.LevelFull)
	return New(Deps{Directory: dir, Links: links, Accounts: accounts, Licensing: &lvl}), dir, accounts, links, &lvl
}

func TestFirstLoginCreatesTheMemberWithGroups(t *testing.T) {
	a, _, accounts, _, _ := setup()
	id, err := a.Authenticate(ctx, "anna", "anna-pw")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(accounts.members[id].groups, []string{"Office"}) {
		t.Fatalf("groups %v", accounts.members[id].groups)
	}
	if again, _ := a.Authenticate(ctx, "anna", "anna-pw"); again != id {
		t.Fatalf("second login %q", again)
	}
}

func TestWrongPasswordIsRefused(t *testing.T) {
	a, _, accounts, _, _ := setup()
	for _, pw := range []string{"wrong", ""} {
		if _, err := a.Authenticate(ctx, "anna", pw); !errors.Is(err, ErrRefused) {
			t.Fatalf("%q: %v", pw, err)
		}
	}
	if len(accounts.members) != 0 {
		t.Fatal("member created on a failed login")
	}
}

func TestUnknownNamesAndAnUnreachableDirectoryFallBackToLocal(t *testing.T) {
	a, dir, _, _, _ := setup()
	if _, err := a.Authenticate(ctx, "oma", "x"); !errors.Is(err, ErrNotInDirectory) {
		t.Fatalf("unknown name: %v", err)
	}
	dir.down = true
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); !errors.Is(err, ErrNotInDirectory) {
		t.Fatalf("down, not linked: %v", err)
	}
}

func TestLinkedMembersAreRefusedWhileTheDirectoryIsDown(t *testing.T) {
	a, dir, _, _, _ := setup()
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); err != nil {
		t.Fatal(err)
	}
	dir.down = true
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); err == nil || errors.Is(err, ErrNotInDirectory) {
		t.Fatalf("got %v", err)
	}
}

func TestRemovedOrDisabledInTheDirectoryLosesAccess(t *testing.T) {
	a, dir, accounts, links, _ := setup()
	id, _ := a.Authenticate(ctx, "anna", "anna-pw")
	delete(dir.users, "anna")
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); !errors.Is(err, ErrRefused) {
		t.Fatalf("removed: %v", err)
	}
	if !accounts.members[id].disabled || !links["anna"].DisabledByDir {
		t.Fatal("not locked out")
	}
	// Back in the directory: enabled again.
	dir.users["anna"] = domain.User{DN: annaDN}
	if got, err := a.Authenticate(ctx, "anna", "anna-pw"); err != nil || got != id || accounts.members[id].disabled {
		t.Fatalf("back: %q %v", got, err)
	}
	// Disabled in Active Directory.
	dir.users["anna"] = domain.User{DN: annaDN, Disabled: true}
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); !errors.Is(err, ErrRefused) || !accounts.members[id].disabled {
		t.Fatalf("AD-disabled: %v", err)
	}
}

func TestMemberGroupsRestrictWhoGetsIn(t *testing.T) {
	a, _, accounts, _, _ := setup()
	a.d.Policy = domain.Policy{MemberGroups: []string{"waylight-users"}}
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); !errors.Is(err, ErrRefused) || len(accounts.members) != 0 {
		t.Fatalf("not in a member group: %v", err)
	}
}

func TestLicenceLevels(t *testing.T) {
	a, _, accounts, _, lvl := setup()
	*lvl = level(licence.LevelNone)
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); !errors.Is(err, ErrNotInDirectory) {
		t.Fatalf("no licence: %v", err)
	}
	*lvl = level(licence.LevelRead)
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); !errors.Is(err, ErrRefused) || len(accounts.members) != 0 {
		t.Fatalf("read-only creates nobody: %v", err)
	}
	*lvl = level(licence.LevelFull)
	id, _ := a.Authenticate(ctx, "anna", "anna-pw")
	*lvl = level(licence.LevelRead)
	if got, err := a.Authenticate(ctx, "anna", "anna-pw"); err != nil || got != id {
		t.Fatalf("read-only keeps existing members: %q %v", got, err)
	}
}

func TestADeletedMemberIsCreatedAgain(t *testing.T) {
	a, _, accounts, _, _ := setup()
	id, _ := a.Authenticate(ctx, "anna", "anna-pw")
	delete(accounts.members, id)
	if again, err := a.Authenticate(ctx, "anna", "anna-pw"); err != nil || accounts.members[again] == nil {
		t.Fatalf("got %q %v", again, err)
	}
}

func TestDirectoryAdminGroupByDNOrName(t *testing.T) {
	dir := &fakeDir{
		users:    map[string]domain.User{"anna": {DN: annaDN, Groups: []string{"cn=IT,ou=groups,dc=example,dc=org"}}},
		password: map[string]string{annaDN: "anna-pw"},
	}
	accounts, lvl := &fakeAccounts{members: map[string]*member{}}, level(licence.LevelFull)
	a := New(Deps{Directory: dir, Links: memLinks{}, Accounts: accounts, Licensing: &lvl, Policy: domain.Policy{AdminGroups: []string{"IT"}}})
	id, err := a.Authenticate(ctx, "anna", "anna-pw")
	if err != nil || !accounts.members[id].admin || accounts.members[id].adminSource != "directory" {
		t.Fatalf("%v %+v", err, accounts.members[id])
	}
	dir.users["anna"] = domain.User{DN: annaDN}
	if _, err := a.Authenticate(ctx, "anna", "anna-pw"); err != nil || accounts.members[id].admin {
		t.Fatalf("after leaving IT: %v %+v", err, accounts.members[id])
	}
}

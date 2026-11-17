package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/shared/kernel"
)

// memLinks is an in-memory AccountLinkRepository.
type memLinks struct{ m map[string]domain.AccountLink }

func (r *memLinks) Save(_ context.Context, l domain.AccountLink) error { r.m[l.ID] = l; return nil }
func (r *memLinks) FindByTokenHash(_ context.Context, h domain.TokenHash) (domain.AccountLink, error) {
	for _, l := range r.m {
		if l.TokenHash == h {
			return l, nil
		}
	}
	return domain.AccountLink{}, domain.ErrInvalidAccountLink
}
func (r *memLinks) List(_ context.Context, p domain.LinkPurpose) ([]domain.AccountLink, error) {
	var out []domain.AccountLink
	for _, l := range r.m {
		if l.Purpose == p {
			out = append(out, l)
		}
	}
	return out, nil
}
func (r *memLinks) Delete(_ context.Context, id string) (bool, error) {
	_, ok := r.m[id]
	delete(r.m, id)
	return ok, nil
}
func (r *memLinks) DeleteForAccount(_ context.Context, account string) error {
	for id, l := range r.m {
		if l.AccountID == account {
			delete(r.m, id)
		}
	}
	return nil
}
func (r *memLinks) DeleteExpired(_ context.Context, now time.Time) error {
	for id, l := range r.m {
		if l.Expired(now) {
			delete(r.m, id)
		}
	}
	return nil
}

func TestInviteGuest(t *testing.T) {
	e := newEnv().withAccount()
	repo := &memLinks{m: map[string]domain.AccountLink{}}
	links := NewAccountLinks(e.deps, repo)
	member := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "m1", Role: "member"})

	if _, _, err := links.CreateInvite(member, InviteInput{Name: "lena", Role: "guest", ValidDays: 7}); !errors.Is(err, domain.ErrOwnerOnly) {
		t.Fatal("only an admin invites", err)
	}
	if _, _, err := links.CreateInvite(ctx, InviteInput{Name: "lena", Role: "member", ValidDays: 7}); !errors.Is(err, domain.ErrMembersUnavailable) {
		t.Fatal("member invitations follow the members switch", err)
	}
	if _, _, err := links.CreateInvite(ctx, InviteInput{Name: "lena", Role: "guest", ValidDays: 31}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("at most 30 days", err)
	}
	inv, token, err := links.CreateInvite(ctx, InviteInput{Name: "lena", Role: "guest", Tabs: []string{"home"}, ValidDays: 7})
	if err != nil || token == "" || inv.ExpiresAt != e.clock.now.AddDate(0, 0, 7) {
		t.Fatalf("%+v %q %v", inv, token, err)
	}
	if repo.m[inv.ID].TokenHash == domain.TokenHash(token) {
		t.Fatal("the token must be stored hashed")
	}
	if _, _, err := links.CreateInvite(ctx, InviteInput{Name: "lena", Role: "guest", ValidDays: 7}); !errors.Is(err, domain.ErrNameTaken) {
		t.Fatal("two open invitations for one name", err)
	}

	info, err := links.Inspect(ctx, token, "c1")
	if err != nil || info.Purpose != domain.LinkInvite || info.Name != "lena" {
		t.Fatalf("inspect %+v %v", info, err)
	}
	// A weak password does not use up the link.
	if _, err := links.Accept(ctx, AcceptLinkRequest{Token: token, Password: "short", ClientKey: "c1"}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("weak password", err)
	}
	if _, err := links.Accept(ctx, AcceptLinkRequest{Token: token, Password: goodPassword, ClientKey: "c1"}); err != nil {
		t.Fatal(err)
	}
	g, err := e.accounts.FindByName(ctx, "lena")
	if err != nil || g.Role() != domain.RoleGuest || len(g.Tabs()) != 1 {
		t.Fatalf("guest not created: %+v %v", g, err)
	}
	if _, err := links.Accept(ctx, AcceptLinkRequest{Token: token, Password: goodPassword, ClientKey: "c1"}); !errors.Is(err, domain.ErrInvalidAccountLink) {
		t.Fatal("a link works once", err)
	}
	if _, _, err := links.CreateInvite(ctx, InviteInput{Name: "lena", Role: "guest", ValidDays: 7}); !errors.Is(err, domain.ErrNameTaken) {
		t.Fatal("name of an existing account", err)
	}
}

func TestInviteExpiresAndCanBeRevoked(t *testing.T) {
	e := newEnv().withAccount()
	e.deps.MembersEnabled = true
	links := NewAccountLinks(e.deps, &memLinks{m: map[string]domain.AccountLink{}})
	_, token, err := links.CreateInvite(ctx, InviteInput{Name: "max", Role: "member", Tabs: []string{"home"}, ValidDays: 1})
	if err != nil {
		t.Fatal(err)
	}
	e.clock.now = e.clock.now.Add(25 * time.Hour)
	if _, err := links.Accept(ctx, AcceptLinkRequest{Token: token, Password: goodPassword, ClientKey: "c"}); !errors.Is(err, domain.ErrInvalidAccountLink) {
		t.Fatal("expired invitation", err)
	}
	if list, _ := links.Invites(ctx); len(list) != 0 {
		t.Fatalf("expired invitations are dropped: %+v", list)
	}

	inv, token, _ := links.CreateInvite(ctx, InviteInput{Name: "max", Role: "member", ValidDays: 3})
	if len(inv.Tabs) != 0 {
		t.Fatal("members get no shared tabs from an invitation")
	}
	if err := links.RevokeInvite(ctx, inv.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := links.Inspect(ctx, token, "c2"); !errors.Is(err, domain.ErrInvalidAccountLink) {
		t.Fatal("revoked invitation", err)
	}
}

func TestResetLink(t *testing.T) {
	e := newEnv().withAccount()
	links := NewAccountLinks(e.deps, &memLinks{m: map[string]domain.AccountLink{}})
	g, err := NewGuests(e.deps).Create(ctx, NewGuestRequest{Name: "oma", Password: goodPassword, Tabs: []string{"home"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewLogin(e.deps).Execute(ctx, LoginRequest{Username: "oma", Password: goodPassword, ClientKey: "x"}); err != nil {
		t.Fatal(err)
	}

	if _, _, err := links.CreateResetLink(ctx, domain.OwnerID); !errors.Is(err, domain.ErrAdminResetByShell) {
		t.Fatal("no reset link for the admin", err)
	}
	old, _, err := links.CreateResetLink(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	token, expires, err := links.CreateResetLink(ctx, g.ID)
	if err != nil || !expires.Equal(e.clock.now.Add(domain.ResetLinkTTL)) {
		t.Fatal(expires, err)
	}
	if _, err := links.Inspect(ctx, old, "c"); !errors.Is(err, domain.ErrInvalidAccountLink) {
		t.Fatal("a new reset link replaces the old one", err)
	}
	const newPassword = "another-good-password"
	info, err := links.Accept(ctx, AcceptLinkRequest{Token: token, Password: newPassword, ClientKey: "c"})
	if err != nil || info.Purpose != domain.LinkReset || info.Name != "oma" {
		t.Fatalf("%+v %v", info, err)
	}
	if active, _ := e.sessions.ListActive(ctx, g.ID, e.clock.now); len(active) != 0 {
		t.Fatal("a reset ends the sessions of the account")
	}
	if _, err := NewLogin(e.deps).Execute(ctx, LoginRequest{Username: "oma", Password: newPassword, ClientKey: "y"}); err != nil {
		t.Fatal("new password", err)
	}
}

func TestAccountLinksAreRateLimited(t *testing.T) {
	e := newEnv().withAccount()
	links := NewAccountLinks(e.deps, &memLinks{m: map[string]domain.AccountLink{}})
	for range 3 {
		if _, err := links.Inspect(ctx, "guess", "c"); !errors.Is(err, domain.ErrInvalidAccountLink) {
			t.Fatal(err)
		}
	}
	if _, err := links.Inspect(ctx, "guess", "c"); !errors.Is(err, domain.ErrTooManyAttempts) {
		t.Fatal("guessing tokens is throttled", err)
	}
}

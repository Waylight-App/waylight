package usecase

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
)

// ownedAccounts stores accounts per owner like the gateway: Get and Save
// only reach the principal's own.
type ownedAccounts struct {
	by map[string]map[string]domain.Account
}

func me(ctx context.Context) string { return kernel.PrincipalFrom(ctx).SettingsAccount() }

func (r *ownedAccounts) List(ctx context.Context) ([]domain.Account, error) {
	var out []domain.Account
	for _, a := range r.by[me(ctx)] {
		out = append(out, a)
	}
	return out, nil
}
func (r *ownedAccounts) Get(ctx context.Context, id string) (domain.Account, error) {
	a, ok := r.by[me(ctx)][id]
	if !ok {
		return a, domain.ErrAccountNotFound
	}
	return a, nil
}
func (r *ownedAccounts) Save(ctx context.Context, a domain.Account) error {
	if _, ok := r.by[me(ctx)][a.ID]; !ok {
		return nil // like the gateway's ON CONFLICT … WHERE owner: silently nothing
	}
	r.by[me(ctx)][a.ID] = a
	return nil
}
func (r *ownedAccounts) Delete(context.Context, string) error { return nil }

// ListSharedAccounts shares every account of "anna" with everyone else.
func (r *ownedAccounts) ListSharedAccounts(ctx context.Context) ([]SharedAccount, error) {
	if me(ctx) == "anna" {
		return nil, nil
	}
	var out []SharedAccount
	for _, a := range r.by["anna"] {
		out = append(out, SharedAccount{Account: a, OwnerID: "anna"})
	}
	return out, nil
}

// rotatingOAuth hands out a new refresh token on every refresh and refuses
// an old one, as Microsoft does.
type rotatingOAuth struct {
	fakeOAuth
	current string
	n       int
}

func (f *rotatingOAuth) Refresh(_ context.Context, _ domain.Provider, _ string, g domain.Grant) (domain.Grant, error) {
	if g.Refresh != f.current {
		return domain.Grant{}, ErrGrantRejected
	}
	f.n++
	f.current = "r" + string(rune('1'+f.n))
	g.Access, g.Refresh, g.Expiry = "a-"+f.current, f.current, now.Add(time.Hour)
	return g, nil
}

func TestASharedAccountIsRefreshedAsItsOwner(t *testing.T) {
	repo := &ownedAccounts{by: map[string]map[string]domain.Account{"anna": {"acc1": {ID: "acc1", Provider: domain.ProviderMicrosoft,
		Grant:     domain.Grant{Access: "a-old", Refresh: "r1", Expiry: now.Add(-time.Minute)},
		Calendars: []domain.AccountCalendar{{ID: "team", Name: "Team", Enabled: true}}}}, "ben": {}}}
	oa := &rotatingOAuth{current: "r1"}
	api := &fakeAPI{}
	uc := NewCalendar(Deps{Repo: &memRepo{}, Source: &fakeSource{}, Writer: &fakeWriter{}, Subscriptions: &memSubs{}, Feeds: &fakeFeeds{},
		IDs: &seqIDs{}, Clock: clk{}, Accounts: repo, API: api, OAuth: oa,
		SharedAccounts: repo, SharingOn: func(context.Context) bool { return true }})

	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	res, err := uc.Upcoming(ben, 7)
	if err != nil || len(res.Events) != 1 || res.Events[0].Calendar != "Team" {
		t.Fatalf("ben's agenda: %+v %v", res, err)
	}
	// The rotated token is stored with anna's account, so it keeps working.
	if got := repo.by["anna"]["acc1"].Grant.Refresh; got != "r2" {
		t.Fatalf("anna's stored refresh token %q, want the rotated r2", got)
	}
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if _, err := uc.Upcoming(anna, 7); err != nil {
		t.Fatalf("anna after ben's refresh: %v", err)
	}
	if res, _ := uc.Upcoming(anna, 7); res.Error != "" {
		t.Fatalf("anna's account broke: %s", res.Error)
	}
}

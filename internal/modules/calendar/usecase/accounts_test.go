package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
)

type memAccounts struct{ m map[string]domain.Account }

func (r *memAccounts) List(context.Context) ([]domain.Account, error) {
	out := []domain.Account{}
	for _, a := range r.m {
		out = append(out, a)
	}
	return out, nil
}
func (r *memAccounts) Get(_ context.Context, id string) (domain.Account, error) {
	a, ok := r.m[id]
	if !ok {
		return a, domain.ErrAccountNotFound
	}
	return a, nil
}
func (r *memAccounts) Save(_ context.Context, a domain.Account) error { r.m[a.ID] = a; return nil }
func (r *memAccounts) Delete(_ context.Context, id string) error {
	if _, ok := r.m[id]; !ok {
		return domain.ErrAccountNotFound
	}
	delete(r.m, id)
	return nil
}

type fakeAPI struct {
	tokens  []string // tokens seen by Events
	created []string // calendarID|summary
	fail    map[string]bool
}

func (f *fakeAPI) Profile(context.Context, domain.Provider, string) (string, error) {
	return "me@example.org", nil
}
func (f *fakeAPI) Calendars(context.Context, domain.Provider, string) ([]domain.AccountCalendar, error) {
	return []domain.AccountCalendar{
		{ID: "primary", Name: "Me", Color: "#112233", Writable: true},
		{ID: "holidays", Name: "Holidays"},
	}, nil
}
func (f *fakeAPI) Events(_ context.Context, _ domain.Provider, token, cal string, _, _ time.Time) ([]domain.Event, error) {
	f.tokens = append(f.tokens, token)
	if f.fail[cal] {
		return nil, errors.New("boom")
	}
	return []domain.Event{{UID: cal + "-1", Summary: "From " + cal, Start: now.Add(2 * time.Hour), End: now.Add(3 * time.Hour)}}, nil
}
func (f *fakeAPI) Create(_ context.Context, _ domain.Provider, _, cal string, e domain.Event) error {
	f.created = append(f.created, cal+"|"+e.Summary)
	return nil
}

type fakeOAuth struct {
	refreshed int
	revoked   []string
	reject    bool
}

func (f *fakeOAuth) RedirectURL() string {
	return "https://home.example/api/calendar/accounts/oauth/callback"
}
func (f *fakeOAuth) NewState() string { return "state-" + time.Now().Format("150405.000000000") }
func (f *fakeOAuth) AuthURL(p domain.Provider, clientID, state, _ string) string {
	return "https://login.example/" + string(p) + "?client_id=" + clientID + "&state=" + state
}
func (f *fakeOAuth) Exchange(_ context.Context, _ domain.Provider, _, secret, code, _ string) (domain.Grant, error) {
	if code == "bad" {
		return domain.Grant{}, errors.New("invalid_grant")
	}
	// Expires right away, so the first use refreshes it.
	return domain.Grant{ClientSecret: secret, Access: "a1", Refresh: "r1", Expiry: now}, nil
}
func (f *fakeOAuth) Refresh(_ context.Context, _ domain.Provider, _ string, g domain.Grant) (domain.Grant, error) {
	if f.reject {
		return domain.Grant{}, ErrGrantRejected
	}
	f.refreshed++
	g.Access, g.Expiry = "a2", now.Add(time.Hour)
	return g, nil
}
func (f *fakeOAuth) Revoke(_ context.Context, _ domain.Provider, _ string, g domain.Grant) error {
	f.revoked = append(f.revoked, g.Refresh)
	return nil
}

func newAccountCalendar(api *fakeAPI, oa *fakeOAuth) (Service, *memAccounts) {
	accounts := &memAccounts{m: map[string]domain.Account{}}
	uc := NewCalendar(Deps{Repo: &memRepo{}, Source: &fakeSource{}, Writer: &fakeWriter{}, Subscriptions: &memSubs{}, Feeds: &fakeFeeds{},
		IDs: &seqIDs{}, Clock: clk{}, Accounts: accounts, API: api, OAuth: oa})
	return uc, accounts
}

func TestAccountSignInAgendaAndCreate(t *testing.T) {
	ctx := context.Background()
	api, oa := &fakeAPI{}, &fakeOAuth{}
	uc, accounts := newAccountCalendar(api, oa)

	if _, err := uc.StartSignIn(ctx, "yahoo", "id", ""); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown provider: %v", err)
	}
	start, err := uc.StartSignIn(ctx, domain.ProviderGoogle, "client-1", "secret")
	if err != nil || start.URL == "" {
		t.Fatalf("%+v %v", start, err)
	}
	if s, _ := uc.SignInStatus(start.State); s != "pending" {
		t.Fatalf("status %q", s)
	}
	if err := uc.FinishSignIn(ctx, "forged", "code", ""); err == nil {
		t.Fatal("an unknown state must be refused")
	}
	if err := uc.FinishSignIn(ctx, start.State, "code", ""); err != nil {
		t.Fatal(err)
	}
	if s, _ := uc.SignInStatus(start.State); s != "ready" {
		t.Fatalf("status %q", s)
	}
	list, _ := uc.ListAccounts(ctx)
	if len(list) != 1 || list[0].Name != "me@example.org" || len(list[0].Calendars) != 2 || !list[0].Calendars[1].Enabled {
		t.Fatalf("%+v", list)
	}
	id := list[0].ID

	// Both calendars are on: their events are in the agenda, with the
	// refreshed token (the first one had already expired).
	agenda, err := uc.Upcoming(ctx, 7)
	if err != nil || len(agenda.Events) != 2 || agenda.Events[0].Color == "" && agenda.Events[1].Color == "" {
		t.Fatalf("%+v %v", agenda, err)
	}
	if oa.refreshed != 1 || api.tokens[len(api.tokens)-1] != "a2" || accounts.m[id].Grant.Access != "a2" {
		t.Fatalf("refresh: %d %v %q", oa.refreshed, api.tokens, accounts.m[id].Grant.Access)
	}

	// Only writable, enabled calendars accept new events.
	w, _ := uc.Writable(ctx)
	if len(w) != 1 || w[0].ID != domain.CalendarRef(id, "primary") || w[0].Name != "Me (Google)" {
		t.Fatalf("writable %+v", w)
	}
	e, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: w[0].ID, Summary: "Dentist", Start: now.Add(time.Hour), End: now.Add(2 * time.Hour)})
	if err != nil || e.Calendar != "Me" || len(api.created) != 1 || api.created[0] != "primary|Dentist" {
		t.Fatalf("create: %+v %v %v", e, err, api.created)
	}
	if _, err := uc.CreateEvent(ctx, NewEventInput{CalendarURL: domain.CalendarRef(id, "holidays"), Summary: "x", Start: now, End: now.Add(time.Hour)}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("read-only calendar: %v", err)
	}

	// Switching a calendar off removes it from the agenda.
	if _, err := uc.SetCalendars(ctx, id, []string{"primary"}); err != nil {
		t.Fatal(err)
	}
	if agenda, _ := uc.Upcoming(ctx, 7); len(agenda.Events) != 1 {
		t.Fatalf("%+v", agenda.Events)
	}

	// Disconnecting revokes the grant and forgets the account.
	if err := uc.Disconnect(ctx, id); err != nil || len(oa.revoked) != 1 || len(accounts.m) != 0 {
		t.Fatalf("%v %v %d", err, oa.revoked, len(accounts.m))
	}
	if _, err := uc.Upcoming(ctx, 7); !errors.Is(err, domain.ErrNotConfigured) {
		t.Fatalf("without sources: %v", err)
	}
}

func TestAccountFailures(t *testing.T) {
	ctx := context.Background()
	api, oa := &fakeAPI{fail: map[string]bool{"holidays": true}}, &fakeOAuth{}
	uc, accounts := newAccountCalendar(api, oa)

	start, _ := uc.StartSignIn(ctx, domain.ProviderMicrosoft, "client", "")
	if err := uc.FinishSignIn(ctx, start.State, "bad", ""); err == nil {
		t.Fatal("a failed exchange must be reported")
	}
	if s, _ := uc.SignInStatus(start.State); s != "failed" {
		t.Fatalf("status %q", s)
	}
	start, _ = uc.StartSignIn(ctx, domain.ProviderMicrosoft, "client", "")
	_ = uc.FinishSignIn(ctx, start.State, "code", "")

	// One failing calendar is reported, the other still shows.
	agenda, err := uc.Upcoming(ctx, 7)
	if err != nil || len(agenda.Events) != 1 || agenda.Error == "" {
		t.Fatalf("%+v %v", agenda, err)
	}

	// A refused refresh asks to sign in again instead of failing silently.
	for id, a := range accounts.m {
		a.Grant.Expiry = now
		accounts.m[id] = a
	}
	oa.reject = true
	res, err := uc.Range(ctx, now.Add(40*24*time.Hour), now.Add(41*24*time.Hour))
	if err == nil && res.Error == "" {
		t.Fatalf("expected the sign-in problem to be reported: %+v", res)
	}

	for range domain.MaxAccounts {
		accounts.m[time.Now().String()] = domain.Account{}
	}
	if _, err := uc.StartSignIn(ctx, domain.ProviderGoogle, "c", ""); !errors.Is(err, kernel.ErrConflict) {
		t.Fatalf("limit: %v", err)
	}
}

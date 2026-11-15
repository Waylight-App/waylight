package usecase

import (
	"context"
	"testing"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
)

// subsByAccount keeps subscriptions per account like the gateway, without
// a fallback to the admin's.
type subsByAccount map[string][]domain.Subscription

func (m subsByAccount) List(ctx context.Context) ([]domain.Subscription, error) {
	return m[kernel.PrincipalFrom(ctx).SettingsAccount()], nil
}
func (m subsByAccount) Get(ctx context.Context, id string) (domain.Subscription, error) {
	for _, s := range m[kernel.PrincipalFrom(ctx).SettingsAccount()] {
		if s.ID == id {
			return s, nil
		}
	}
	return domain.Subscription{}, domain.ErrSubscriptionNotFound
}
func (m subsByAccount) Save(ctx context.Context, s domain.Subscription) error {
	a := kernel.PrincipalFrom(ctx).SettingsAccount()
	m[a] = append(m[a], s)
	return nil
}
func (m subsByAccount) Delete(context.Context, string) error { return nil }

// feedNamedEvent returns one event named after the subscription.
type feedNamedEvent struct{}

func (feedNamedEvent) Events(_ context.Context, s domain.Subscription, from, _ time.Time) ([]domain.Event, error) {
	start := from.Add(48 * time.Hour)
	return []domain.Event{{UID: s.ID, Summary: s.Name, Start: start, End: start.Add(time.Hour)}}, nil
}

// TestAgendaPerAccount: every account sees its own calendars; members do
// not see the admin's, guests do, and the job refreshes everyone (ADR 0018).
func TestAgendaPerAccount(t *testing.T) {
	subs := subsByAccount{
		kernel.SystemAccountID: {{ID: "s-admin", Name: "Admin dentist", URL: "https://example.org/a.ics", RefreshInterval: time.Hour}},
		"anna":                 {{ID: "s-anna", Name: "Anna's yoga", URL: "https://example.org/b.ics", RefreshInterval: time.Hour}},
	}
	uc := NewCalendar(Deps{Repo: &memRepo{}, Source: &fakeSource{}, Writer: &fakeWriter{}, Subscriptions: subs, Feeds: feedNamedEvent{},
		IDs: &seqIDs{}, Clock: clk{}})
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})

	summaries := func(ctx context.Context) []string {
		res, err := uc.Upcoming(ctx, 7)
		if err != nil {
			return []string{"error: " + err.Error()}
		}
		var out []string
		for _, e := range res.Events {
			out = append(out, e.Summary)
		}
		return out
	}
	if got := summaries(anna); len(got) != 1 || got[0] != "Anna's yoga" {
		t.Fatalf("anna sees %v", got)
	}
	if got := summaries(guest); len(got) != 1 || got[0] != "Admin dentist" {
		t.Fatalf("the guest sees %v", got)
	}
	if got := summaries(ben); len(got) != 1 || got[0][:6] != "error:" {
		t.Fatalf("ben has no calendars and must not see the admin's: %v", got)
	}
	// A subscription anna adds later reaches her agenda through the job.
	subs["anna"] = append(subs["anna"], domain.Subscription{ID: "s-anna2", Name: "Anna's choir", URL: "https://example.org/c.ics", RefreshInterval: time.Hour})
	if err := uc.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := summaries(anna); len(got) != 2 {
		t.Fatalf("the job did not refresh anna's agenda: %v", got)
	}
	if got := summaries(guest); len(got) != 1 {
		t.Fatalf("anna's calendars leaked to the guest: %v", got)
	}
}

// ownerRecorder remembers for which account each connected account was saved.
type ownerRecorder struct {
	memAccounts
	savedFor []string
}

func (r *ownerRecorder) Save(ctx context.Context, a domain.Account) error {
	r.savedFor = append(r.savedFor, kernel.PrincipalFrom(ctx).SettingsAccount())
	return r.memAccounts.Save(ctx, a)
}

// TestSignInConnectsTheStartingAccount: the OAuth callback arrives without
// a session; the account must be stored for the member who started.
func TestSignInConnectsTheStartingAccount(t *testing.T) {
	rec := &ownerRecorder{memAccounts: memAccounts{m: map[string]domain.Account{}}}
	uc := NewCalendar(Deps{Repo: &memRepo{}, Source: &fakeSource{}, Writer: &fakeWriter{}, Subscriptions: &memSubs{}, Feeds: &fakeFeeds{},
		IDs: &seqIDs{}, Clock: clk{}, Accounts: rec, API: &fakeAPI{}, OAuth: &fakeOAuth{}})
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	start, err := uc.StartSignIn(anna, domain.ProviderGoogle, "client-1", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.FinishSignIn(context.Background(), start.State, "code", ""); err != nil {
		t.Fatal(err)
	}
	if len(rec.savedFor) == 0 || rec.savedFor[0] != "anna" {
		t.Fatalf("the connected account was saved for %v", rec.savedFor)
	}
}

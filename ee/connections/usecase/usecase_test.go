package usecase

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"trisse-home/ee/connections/domain"
	licence "trisse-home/ee/licence/domain"
	"trisse-home/internal/shared/kernel"
)

type memTokens map[string]domain.Connection

func key(a, s string) string { return a + "|" + s }

func (m memTokens) Get(_ context.Context, a, s string) (domain.Connection, bool, error) {
	c, ok := m[key(a, s)]
	return c, ok, nil
}
func (m memTokens) ListFor(_ context.Context, a string) ([]domain.Connection, error) {
	var out []domain.Connection
	for _, c := range m {
		if c.AccountID == a {
			out = append(out, c)
		}
	}
	return out, nil
}
func (m memTokens) Save(_ context.Context, c domain.Connection) error {
	m[key(c.AccountID, c.Service)] = c
	return nil
}
func (m memTokens) Delete(_ context.Context, a, s string) error { delete(m, key(a, s)); return nil }
func (m memTokens) DeleteAccount(_ context.Context, a string) error {
	for k, c := range m {
		if c.AccountID == a {
			delete(m, k)
		}
	}
	return nil
}

type memClients map[string]domain.Client

func (m memClients) Get(_ context.Context, s string) (domain.Client, bool, error) {
	c, ok := m[s]
	return c, ok, nil
}
func (m memClients) List(context.Context) ([]domain.Client, error) { return nil, nil }
func (m memClients) Save(_ context.Context, c domain.Client) error { m[c.Service] = c; return nil }

// fakeOAuth issues tokens for code "good"; refresh tokens can be revoked.
type fakeOAuth struct {
	verifiers map[string]string // state -> verifier, captured from AuthCodeURL
	revoked   map[string]bool
	refreshes int
	now       func() time.Time
}

func (f *fakeOAuth) Services() []ServiceInfo {
	return []ServiceInfo{{ID: "notion", Name: "Notion"}, {ID: "microsoft", Name: "Microsoft 365"}}
}
func (f *fakeOAuth) RedirectURL() string { return "https://home.example/api/connections/callback" }
func (f *fakeOAuth) AuthCodeURL(c domain.Client, state, verifier string) (string, error) {
	f.verifiers[state] = verifier
	return "https://id.example/auth?" + url.Values{"state": {state}, "client_id": {c.ClientID}}.Encode(), nil
}
func (f *fakeOAuth) Exchange(_ context.Context, _ domain.Client, code, verifier string) (domain.Token, error) {
	if code != "good" || verifier == "" {
		return domain.Token{}, errors.New("invalid_grant")
	}
	return domain.Token{Access: "access-1", Refresh: "refresh-1", Expiry: f.now().Add(time.Hour)}, nil
}
func (f *fakeOAuth) Refresh(_ context.Context, _ domain.Client, refresh string) (domain.Token, error) {
	f.refreshes++
	if f.revoked[refresh] {
		return domain.Token{}, ErrRejected
	}
	return domain.Token{Access: "access-2", Refresh: refresh, Expiry: f.now().Add(time.Hour)}, nil
}
func (f *fakeOAuth) Revoke(_ context.Context, _ domain.Client, tok string) error {
	f.revoked[tok] = true
	return nil
}

type clock struct{ t time.Time }

func (c *clock) Now() time.Time { return c.t }

type lvl licence.Level

func (l lvl) Level(context.Context, string) licence.Level { return licence.Level(l) }

var (
	anna  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "m1", Role: "member"})
	ben   = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "m2", Role: "member"})
	admin = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner"})
)

func setup() (*Service, memTokens, *fakeOAuth, *clock, *[]string) {
	c := &clock{t: time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)}
	f := &fakeOAuth{verifiers: map[string]string{}, revoked: map[string]bool{}, now: c.Now}
	tokens := memTokens{}
	var audit []string
	s := New(Deps{Tokens: tokens, Clients: memClients{"notion": {Service: "notion", ClientID: "cid", Secret: "sec"}}, OAuth: f, Clock: c,
		Licensing: lvl(licence.LevelFull), Audit: func(_ context.Context, action, service string) { audit = append(audit, action+" "+service) }})
	return s, tokens, f, c, &audit
}

func stateOf(t *testing.T, u string) string {
	t.Helper()
	parsed, _ := url.Parse(u)
	return parsed.Query().Get("state")
}

func TestConnectRefreshAndDisconnect(t *testing.T) {
	s, tokens, f, c, audit := setup()
	u, err := s.Start(anna, "notion")
	if err != nil {
		t.Fatal(err)
	}
	if svc, err := s.Finish(anna, stateOf(t, u), "good", ""); err != nil || svc != "notion" {
		t.Fatalf("finish: %q %v", svc, err)
	}
	if tok, err := s.AccessToken(anna, "m1", "notion"); err != nil || tok != "access-1" {
		t.Fatalf("token: %q %v", tok, err)
	}
	c.t = c.t.Add(59 * time.Minute) // within the refresh margin
	if tok, err := s.AccessToken(anna, "m1", "notion"); err != nil || tok != "access-2" || f.refreshes != 1 {
		t.Fatalf("refreshed token: %q %v (%d refreshes)", tok, err, f.refreshes)
	}
	if err := s.Disconnect(anna, "notion"); err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 0 || !f.revoked["refresh-1"] {
		t.Fatalf("after disconnect: %v, revoked %v", tokens, f.revoked)
	}
	if strings.Join(*audit, ",") != "connected notion,disconnected notion" {
		t.Fatalf("audit: %v", *audit)
	}
}

func TestATamperedCallbackIsRefused(t *testing.T) {
	s, tokens, _, c, _ := setup()
	u, _ := s.Start(anna, "notion")
	state := stateOf(t, u)
	if _, err := s.Finish(anna, "forged-state", "good", ""); !errors.Is(err, ErrSignInExpired) {
		t.Fatalf("forged state: %v", err)
	}
	// Someone else's browser cannot finish anna's sign-in.
	if _, err := s.Finish(ben, state, "good", ""); !errors.Is(err, ErrSignInExpired) {
		t.Fatalf("other account: %v", err)
	}
	// The state was used up by that attempt.
	if _, err := s.Finish(anna, state, "good", ""); !errors.Is(err, ErrSignInExpired) {
		t.Fatalf("state reused: %v", err)
	}
	u, _ = s.Start(anna, "notion")
	c.t = c.t.Add(11 * time.Minute)
	if _, err := s.Finish(anna, stateOf(t, u), "good", ""); !errors.Is(err, ErrSignInExpired) {
		t.Fatalf("expired: %v", err)
	}
	u, _ = s.Start(anna, "notion")
	if _, err := s.Finish(anna, stateOf(t, u), "bad-code", ""); err == nil {
		t.Fatal("a code the service refuses was accepted")
	}
	if len(tokens) != 0 {
		t.Fatalf("tokens stored: %v", tokens)
	}
}

func TestARevokedGrantAsksToReconnect(t *testing.T) {
	s, tokens, f, c, _ := setup()
	u, _ := s.Start(anna, "notion")
	_, _ = s.Finish(anna, stateOf(t, u), "good", "")
	f.revoked["refresh-1"] = true // revoked at the service
	c.t = c.t.Add(2 * time.Hour)
	if _, err := s.AccessToken(anna, "m1", "notion"); !errors.Is(err, ErrReconnect) {
		t.Fatalf("got %v", err)
	}
	if tokens[key("m1", "notion")].Status != domain.StatusExpired {
		t.Fatal("status not expired")
	}
	views, _ := s.List(anna)
	if views[0].Status != domain.StatusExpired {
		t.Fatalf("list: %+v", views)
	}
}

func TestStartNeedsARegisteredClient(t *testing.T) {
	s, _, _, _, _ := setup()
	if _, err := s.Start(anna, "microsoft"); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("unconfigured: %v", err)
	}
	if _, err := s.Start(anna, "myspace"); !errors.Is(err, ErrUnknownService) {
		t.Fatalf("unknown: %v", err)
	}
}

func TestClientsAreForAdminsAndKeepTheSecret(t *testing.T) {
	s, _, _, _, _ := setup()
	if err := s.SaveClient(anna, domain.Client{Service: "notion", ClientID: "x"}); err == nil {
		t.Fatal("member registered a client")
	}
	if err := s.SaveClient(admin, domain.Client{Service: "notion", ClientID: "new-id"}); err != nil {
		t.Fatal(err)
	}
	views, redirect, err := s.ClientsList(admin)
	if err != nil || redirect == "" || views[0].ClientID != "new-id" || !views[0].HasSecret {
		t.Fatalf("clients: %+v %q %v", views, redirect, err)
	}
}

func TestDeletedAccountsLoseTheirTokens(t *testing.T) {
	s, tokens, _, _, _ := setup()
	u, _ := s.Start(anna, "notion")
	_, _ = s.Finish(anna, stateOf(t, u), "good", "")
	if err := s.AccountDeleted(context.Background(), "m1"); err != nil || len(tokens) != 0 {
		t.Fatalf("tokens left: %v %v", tokens, err)
	}
}

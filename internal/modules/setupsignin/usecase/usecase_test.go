package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/setupsignin/domain"
	"waylight/internal/shared/kernel"
)

type clock struct{ now time.Time }

func (c *clock) Now() time.Time { return c.now }

// fakeNextcloud grants access after the user "signed in" (done) and counts
// polls.
type fakeNextcloud struct {
	startErr error
	pollErr  error
	login    string
	done     bool
	polls    int
	gotBase  string
}

func (f *fakeNextcloud) Start(_ context.Context, base string) (string, string, error) {
	f.gotBase = base
	if f.startErr != nil {
		return "", "", f.startErr
	}
	login := f.login
	if login == "" {
		login = base + "/login/v2/flow/abc"
	}
	return login, "poll-token", nil
}

func (f *fakeNextcloud) Poll(_ context.Context, base, token string) (string, string, bool, error) {
	f.polls++
	if token != "poll-token" || base != f.gotBase {
		return "", "", false, errors.New("wrong poll")
	}
	if f.pollErr != nil {
		return "", "", false, f.pollErr
	}
	if !f.done {
		return "", "", false, nil
	}
	return "anna", "app-password", true, nil
}

type fakeHA struct {
	clientID string
	codes    []string
	err      error
}

func (f *fakeHA) ClientID() string { return f.clientID }
func (f *fakeHA) AuthURL(base, state string) string {
	return base + "/auth/authorize?state=" + state
}
func (f *fakeHA) Exchange(_ context.Context, _, code string) (kernel.SignInGrant, error) {
	f.codes = append(f.codes, code)
	if f.err != nil {
		return kernel.SignInGrant{}, f.err
	}
	return kernel.SignInGrant{Access: "access", Secret: "refresh", ClientID: f.clientID}, nil
}

type fixture struct {
	uc    SignIns
	nc    *fakeNextcloud
	ha    *fakeHA
	clock *clock
}

func setup() fixture {
	f := fixture{nc: &fakeNextcloud{}, ha: &fakeHA{clientID: "https://home.example/"}, clock: &clock{now: time.Unix(1_700_000_000, 0)}}
	n := 0
	f.uc = New(Deps{Nextcloud: f.nc, HomeAssistant: f.ha, Clock: f.clock, Token: func() string {
		n++
		return fmt.Sprintf("token-%d", n)
	}})
	return f
}

func as(id, role string) context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: id, Role: role})
}

// stateOf reads the state from the fake authorise URL.
func stateOf(t *testing.T, login string) string {
	t.Helper()
	_, state, ok := strings.Cut(login, "state=")
	if !ok || state == "" {
		t.Fatalf("no state in %q", login)
	}
	return state
}

// TestNextcloudFlow: start, poll no more often than PollInterval, finish,
// redeem once and only as the account that started it.
func TestNextcloudFlow(t *testing.T) {
	f := setup()
	anna, ben := as("anna", "admin"), as("ben", "member")
	s, err := f.uc.Start(anna, "nextcloud", " https://cloud.example/ ")
	if err != nil || s.LoginURL != "https://cloud.example/login/v2/flow/abc" || f.nc.gotBase != "https://cloud.example" {
		t.Fatalf("start: %+v %v", s, err)
	}
	if st, err := f.uc.Status(anna, s.ID); err != nil || st != domain.Pending || f.nc.polls != 1 {
		t.Fatalf("first status: %v %v polls=%d", st, err, f.nc.polls)
	}
	if _, err := f.uc.Status(anna, s.ID); err != nil || f.nc.polls != 1 {
		t.Fatalf("polled again within the interval: %d", f.nc.polls)
	}
	if _, err := f.uc.Status(ben, s.ID); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("another account must not follow the sign-in: %v", err)
	}
	if _, err := f.uc.Redeem(anna, s.ID, kernel.SignInNextcloud); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("an unfinished sign-in cannot be redeemed: %v", err)
	}
	f.nc.done = true
	f.clock.now = f.clock.now.Add(domain.PollInterval)
	if st, err := f.uc.Status(anna, s.ID); err != nil || st != domain.Ready {
		t.Fatalf("ready: %v %v", st, err)
	}
	if _, err := f.uc.Redeem(ben, s.ID, kernel.SignInNextcloud); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("another account must not redeem: %v", err)
	}
	if _, err := f.uc.Redeem(anna, s.ID, kernel.SignInHomeAssistant); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("redeemed as the wrong service: %v", err)
	}
	g, err := f.uc.Redeem(anna, s.ID, kernel.SignInNextcloud)
	if err != nil || g.URL != "https://cloud.example" || g.Username != "anna" || g.Secret != "app-password" {
		t.Fatalf("redeem: %+v %v", g, err)
	}
	if _, err := f.uc.Redeem(anna, s.ID, kernel.SignInNextcloud); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("redeemed twice: %v", err)
	}
}

// TestNextcloudFailures: a wrong address fails at the start with a setup
// message, repeated poll failures end the sign-in, and it expires.
func TestNextcloudFailures(t *testing.T) {
	f := setup()
	anna := as("anna", "admin")
	f.nc.startErr = kernel.ErrTargetWrongService
	if _, err := f.uc.Start(anna, "nextcloud", "https://wrong.example"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("wrong service: %v", err)
	}
	f.nc.startErr = nil
	if _, err := f.uc.Start(anna, "nextcloud", "javascript:alert(1)"); !errors.Is(err, domain.ErrURL) {
		t.Fatalf("address: %v", err)
	}
	f.nc.login = "javascript:alert(1)"
	if _, err := f.uc.Start(anna, "nextcloud", "https://cloud.example"); !errors.Is(err, errBadPage) {
		t.Fatalf("a login page that is not http(s): %v", err)
	}
	f.nc.login = ""

	s, _ := f.uc.Start(anna, "nextcloud", "https://cloud.example")
	f.nc.pollErr = errors.New("connection refused")
	var err error
	for range domain.MaxPollFailures {
		_, err = f.uc.Status(anna, s.ID)
		f.clock.now = f.clock.now.Add(domain.PollInterval)
	}
	if !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("repeated poll failures end the sign-in: %v", err)
	}

	f.nc.pollErr = nil
	s, _ = f.uc.Start(anna, "nextcloud", "https://cloud.example")
	f.clock.now = f.clock.now.Add(domain.TTL)
	if _, err := f.uc.Status(anna, s.ID); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("expired: %v", err)
	}
}

// TestHomeAssistantFlow: the callback is tied to the sign-in by its state,
// works once, and the grant goes to the account that started it.
func TestHomeAssistantFlow(t *testing.T) {
	f := setup()
	anna := as("anna", "admin")
	s, err := f.uc.Start(anna, "homeassistant", "http://ha.local:8123/")
	if err != nil {
		t.Fatal(err)
	}
	state := stateOf(t, s.LoginURL)
	if state == s.ID {
		t.Fatal("the state must differ from the ID the browser polls with")
	}
	if err := f.uc.Callback(context.Background(), "forged", "code", ""); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("forged state: %v", err)
	}
	if st, _ := f.uc.Status(anna, s.ID); st != domain.Pending {
		t.Fatalf("a forged callback must not touch the sign-in: %v", st)
	}
	// The redirect arrives without a session.
	if err := f.uc.Callback(context.Background(), state, "code-1", ""); err != nil {
		t.Fatal(err)
	}
	if err := f.uc.Callback(context.Background(), state, "code-2", ""); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("replayed callback: %v", err)
	}
	if len(f.ha.codes) != 1 || f.ha.codes[0] != "code-1" {
		t.Fatalf("exchanged %v", f.ha.codes)
	}
	if st, err := f.uc.Status(anna, s.ID); err != nil || st != domain.Ready {
		t.Fatalf("ready: %v %v", st, err)
	}
	g, err := f.uc.Redeem(anna, s.ID, kernel.SignInHomeAssistant)
	if err != nil || g.URL != "http://ha.local:8123" || g.Secret != "refresh" || g.ClientID != "https://home.example/" {
		t.Fatalf("redeem: %+v %v", g, err)
	}
}

func TestHomeAssistantFailures(t *testing.T) {
	f := setup()
	anna := as("anna", "admin")
	f.ha.clientID = ""
	if _, err := f.uc.Start(anna, "homeassistant", "http://ha.local:8123"); !errors.Is(err, errNoPublicURL) {
		t.Fatalf("without PUBLIC_URL: %v", err)
	}
	f.ha.clientID = "https://home.example/"

	s, _ := f.uc.Start(anna, "homeassistant", "http://ha.local:8123")
	if err := f.uc.Callback(context.Background(), stateOf(t, s.LoginURL), "", "access_denied"); !errors.Is(err, errDenied) {
		t.Fatalf("denied: %v", err)
	}
	if _, err := f.uc.Status(anna, s.ID); !errors.Is(err, errDenied) {
		t.Fatalf("the assistant learns why: %v", err)
	}

	s, _ = f.uc.Start(anna, "homeassistant", "http://ha.local:8123")
	f.ha.err = kernel.ErrTargetAuth
	if err := f.uc.Callback(context.Background(), stateOf(t, s.LoginURL), "code", ""); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("refused code: %v", err)
	}
	if _, err := f.uc.Redeem(anna, s.ID, kernel.SignInHomeAssistant); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("a failed sign-in cannot be redeemed: %v", err)
	}
}

// TestRules: who may sign in, what a start replaces, cancel and the limit.
func TestRules(t *testing.T) {
	f := setup()
	for _, role := range []string{"guest", "kiosk", "token"} {
		if _, err := f.uc.Start(as("x", role), "nextcloud", "https://cloud.example"); !errors.Is(err, kernel.ErrForbidden) {
			t.Errorf("%s: %v", role, err)
		}
	}
	anna := as("anna", "admin")
	if _, err := f.uc.Start(anna, "grafana", "https://grafana.example"); !errors.Is(err, domain.ErrUnknownService) {
		t.Fatalf("unknown service: %v", err)
	}
	first, _ := f.uc.Start(anna, "nextcloud", "https://cloud.example")
	second, _ := f.uc.Start(anna, "nextcloud", "https://cloud.example")
	if _, err := f.uc.Status(anna, first.ID); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("a new start replaces the earlier sign-in: %v", err)
	}
	if err := f.uc.Cancel(as("ben", "member"), second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.uc.Status(anna, second.ID); err != nil {
		t.Fatalf("another account cannot cancel: %v", err)
	}
	if err := f.uc.Cancel(anna, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.uc.Status(anna, second.ID); !errors.Is(err, kernel.ErrSignInUnknown) {
		t.Fatalf("cancelled: %v", err)
	}
	for i := range domain.MaxPending {
		if _, err := f.uc.Start(as(fmt.Sprint("user", i), "member"), "nextcloud", "https://cloud.example"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.uc.Start(anna, "nextcloud", "https://cloud.example"); !errors.Is(err, kernel.ErrTooManyRequests) {
		t.Fatalf("limit: %v", err)
	}
}

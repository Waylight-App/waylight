package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/calendar/domain"
	"trisse-home/internal/shared/kernel"
)

// AccountRepository stores connected accounts with encrypted tokens
// (output port, #282).
type AccountRepository interface {
	List(ctx context.Context) ([]domain.Account, error)
	Get(ctx context.Context, id string) (domain.Account, error)
	Save(ctx context.Context, a domain.Account) error
	Delete(ctx context.Context, id string) error
}

// AccountAPI talks to the calendar API of a provider (output port).
type AccountAPI interface {
	// Profile returns the e-mail address or name of the signed-in account.
	Profile(ctx context.Context, p domain.Provider, token string) (string, error)
	Calendars(ctx context.Context, p domain.Provider, token string) ([]domain.AccountCalendar, error)
	Events(ctx context.Context, p domain.Provider, token, calendarID string, from, to time.Time) ([]domain.Event, error)
	Create(ctx context.Context, p domain.Provider, token, calendarID string, e domain.Event) error
}

// OAuth signs in with the user's own app registration (output port); the
// module wires it to the platform's OAuth client, as for to-dos (#171).
type OAuth interface {
	// RedirectURL is the address to register at the provider; "" without
	// PUBLIC_URL, in which case signing in is not available.
	RedirectURL() string
	NewState() string
	AuthURL(p domain.Provider, clientID, state, verifier string) string
	Exchange(ctx context.Context, p domain.Provider, clientID, clientSecret, code, verifier string) (domain.Grant, error)
	// Refresh fails with an error wrapping ErrGrantRejected when the user
	// has to sign in again.
	Refresh(ctx context.Context, p domain.Provider, clientID string, g domain.Grant) (domain.Grant, error)
	Revoke(ctx context.Context, p domain.Provider, clientID string, g domain.Grant) error
}

// ErrGrantRejected is wrapped by OAuth.Refresh for refused grants.
var ErrGrantRejected = errors.New("grant rejected")

// AccountView is an account without tokens.
type AccountView struct {
	ID        string
	Provider  domain.Provider
	Name      string
	Calendars []domain.AccountCalendar
}

// OAuthStart is where the browser signs in.
type OAuthStart struct {
	URL   string
	State string
}

// CalendarChoice is a calendar new events can be written to.
type CalendarChoice struct {
	// ID is a CalDAV collection URL or a domain.CalendarRef.
	ID   string
	Name string
}

// Accounts connects Google and Microsoft accounts (input port, #282).
type Accounts interface {
	// RedirectURL is what to register at the provider ("" = unavailable).
	RedirectURL() string
	StartSignIn(ctx context.Context, p domain.Provider, clientID, clientSecret string) (OAuthStart, error)
	// FinishSignIn is called by the provider's redirect; on success the
	// account is stored with all its calendars switched on.
	FinishSignIn(ctx context.Context, state, code, providerError string) error
	// SignInStatus is "pending", "ready" or "failed".
	SignInStatus(state string) (string, error)
	ListAccounts(ctx context.Context) ([]AccountView, error)
	// SetCalendars chooses which calendars of an account are shown.
	SetCalendars(ctx context.Context, accountID string, enabled []string) (AccountView, error)
	// Disconnect revokes the tokens at the provider and forgets the account.
	Disconnect(ctx context.Context, accountID string) error
	// Writable lists every calendar that accepts new events.
	Writable(ctx context.Context) ([]CalendarChoice, error)
}

// pendingSignIn is a sign-in between StartSignIn and the redirect.
type pendingSignIn struct {
	// account started the sign-in; the callback arrives without a session,
	// so the connected account is stored for it (ADR 0018).
	account      string
	provider     domain.Provider
	clientID     string
	clientSecret string
	verifier     string
	started      time.Time
	done         bool
	err          error
}

// Limits of sign-ins that are in progress (kept in memory only).
const (
	signInTTL         = 15 * time.Minute
	maxPendingSignIns = 20
	maxClientIDLength = 200
	maxSecretLength   = 500
	// refreshMargin renews tokens shortly before they expire.
	refreshMargin = 2 * time.Minute
)

var (
	errSignInUnknown = kernel.NewPublicError(kernel.ErrNotFound, "The sign-in was not started here or has expired. Please start again.")
	errNoPublicURL   = kernel.NewValidationError("provider", "Signing in needs PUBLIC_URL (an https:// address of trisse-home).")
	errClientID      = kernel.NewValidationError("clientId", "Enter the client ID of your app registration.")
	errTooManySignIn = kernel.NewValidationError("provider", "Too many sign-ins at once. Please wait a few minutes.")
)

type accountState struct {
	mu      sync.Mutex
	pending map[string]*pendingSignIn
	// tokens serialises refreshes, so two requests never use up one
	// refresh token at the same time.
	tokens sync.Mutex
}

func (uc *calendar) RedirectURL() string {
	if uc.oauth == nil {
		return ""
	}
	return uc.oauth.RedirectURL()
}

func (uc *calendar) StartSignIn(ctx context.Context, p domain.Provider, clientID, clientSecret string) (OAuthStart, error) {
	if !p.Valid() {
		return OAuthStart{}, domain.ErrUnknownProvider
	}
	if uc.RedirectURL() == "" {
		return OAuthStart{}, errNoPublicURL
	}
	if clientID == "" || len(clientID) > maxClientIDLength || len(clientSecret) > maxSecretLength {
		return OAuthStart{}, errClientID
	}
	list, err := uc.accounts.List(ctx)
	if err != nil {
		return OAuthStart{}, fmt.Errorf("list accounts: %w", err)
	}
	if len(list) >= domain.MaxAccounts {
		return OAuthStart{}, domain.ErrTooManyAccounts
	}
	state, verifier := uc.oauth.NewState(), uc.oauth.NewState()
	now := uc.clock.Now()
	uc.acc.mu.Lock()
	defer uc.acc.mu.Unlock()
	for k, s := range uc.acc.pending {
		if now.Sub(s.started) > signInTTL {
			delete(uc.acc.pending, k)
		}
	}
	if len(uc.acc.pending) >= maxPendingSignIns {
		return OAuthStart{}, errTooManySignIn
	}
	uc.acc.pending[state] = &pendingSignIn{account: kernel.PrincipalFrom(ctx).SettingsAccount(), provider: p, clientID: clientID, clientSecret: clientSecret, verifier: verifier, started: now}
	return OAuthStart{URL: uc.oauth.AuthURL(p, clientID, state, verifier), State: state}, nil
}

func (uc *calendar) FinishSignIn(ctx context.Context, state, code, providerError string) error {
	uc.acc.mu.Lock()
	p := uc.acc.pending[state]
	uc.acc.mu.Unlock()
	if p == nil || state == "" {
		return errSignInUnknown
	}
	ctx = kernel.WithPrincipal(ctx, kernel.PrincipalFor(p.account))
	err := uc.connect(ctx, p, code, providerError)
	uc.acc.mu.Lock()
	p.done, p.err = true, err
	uc.acc.mu.Unlock()
	if err != nil {
		return kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	_ = uc.Refresh(ctx) // the new calendars show up right away
	return nil
}

// connect exchanges the code and stores the account with its calendars.
func (uc *calendar) connect(ctx context.Context, p *pendingSignIn, code, providerError string) error {
	if providerError != "" || code == "" {
		return fmt.Errorf("provider answered %q", providerError)
	}
	g, err := uc.oauth.Exchange(ctx, p.provider, p.clientID, p.clientSecret, code, p.verifier)
	if err != nil {
		return err
	}
	name, err := uc.api.Profile(ctx, p.provider, g.Access)
	if err != nil {
		return fmt.Errorf("read profile: %w", err)
	}
	cals, err := uc.api.Calendars(ctx, p.provider, g.Access)
	if err != nil {
		return fmt.Errorf("list calendars: %w", err)
	}
	a := domain.Account{ID: uc.ids.NewID(), Provider: p.provider, ClientID: p.clientID, Name: name, Grant: g}.MergeCalendars(cals)
	return uc.accounts.Save(ctx, a)
}

func (uc *calendar) SignInStatus(state string) (string, error) {
	uc.acc.mu.Lock()
	defer uc.acc.mu.Unlock()
	p := uc.acc.pending[state]
	switch {
	case p == nil:
		return "", errSignInUnknown
	case !p.done:
		return "pending", nil
	case p.err != nil:
		return "failed", nil
	}
	return "ready", nil
}

func view(a domain.Account) AccountView {
	return AccountView{ID: a.ID, Provider: a.Provider, Name: a.Name, Calendars: a.Calendars}
}

func (uc *calendar) ListAccounts(ctx context.Context) ([]AccountView, error) {
	list, err := uc.accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	out := make([]AccountView, 0, len(list))
	for _, a := range list {
		out = append(out, view(a))
	}
	return out, nil
}

func (uc *calendar) SetCalendars(ctx context.Context, accountID string, enabled []string) (AccountView, error) {
	a, err := uc.accounts.Get(ctx, accountID)
	if err != nil {
		return AccountView{}, err
	}
	a = a.WithEnabled(enabled)
	if err := uc.accounts.Save(ctx, a); err != nil {
		return AccountView{}, fmt.Errorf("save account: %w", err)
	}
	_ = uc.Refresh(ctx)
	return view(a), nil
}

func (uc *calendar) Disconnect(ctx context.Context, accountID string) error {
	a, err := uc.accounts.Get(ctx, accountID)
	if err != nil {
		return err
	}
	// The account is forgotten even if the provider cannot be reached; the
	// grant then stays valid there until the user revokes it in the
	// provider's account settings.
	_ = uc.oauth.Revoke(ctx, a.Provider, a.ClientID, a.Grant)
	if err := uc.accounts.Delete(ctx, accountID); err != nil {
		return err
	}
	_ = uc.Refresh(ctx)
	return nil
}

func (uc *calendar) Writable(ctx context.Context) ([]CalendarChoice, error) {
	out := []CalendarChoice{}
	if s, err := uc.repo.Load(ctx); err == nil {
		for _, u := range s.CalendarURLs {
			out = append(out, CalendarChoice{ID: u, Name: domain.CalendarName(u)})
		}
	} else if !errors.Is(err, domain.ErrNotConfigured) {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	list, err := uc.accounts.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	for _, a := range list {
		for _, c := range a.Enabled() {
			if c.Writable {
				out = append(out, CalendarChoice{ID: domain.CalendarRef(a.ID, c.ID), Name: c.Name + " (" + a.Provider.Label() + ")"})
			}
		}
	}
	return out, nil
}

// accessToken returns a valid access token, refreshing and storing it when
// it is about to expire.
func (uc *calendar) accessToken(ctx context.Context, a domain.Account) (string, error) {
	uc.acc.tokens.Lock()
	defer uc.acc.tokens.Unlock()
	// Another request may have refreshed it meanwhile.
	if stored, err := uc.accounts.Get(ctx, a.ID); err == nil {
		a = stored
	}
	g := a.Grant
	if g.Expiry.IsZero() || g.Refresh == "" || uc.clock.Now().Add(refreshMargin).Before(g.Expiry) {
		return g.Access, nil
	}
	ng, err := uc.oauth.Refresh(ctx, a.Provider, a.ClientID, g)
	if errors.Is(err, ErrGrantRejected) {
		return "", domain.ErrSignInAgain
	}
	if err != nil {
		return "", err
	}
	a.Grant = ng
	if err := uc.accounts.Save(ctx, a); err != nil {
		return "", fmt.Errorf("store refreshed token: %w", err)
	}
	return ng.Access, nil
}

// accountEvents loads the events of every enabled calendar of the accounts;
// failures are returned per calendar so one broken account does not hide
// the others.
func (uc *calendar) accountEvents(ctx context.Context, accounts []domain.Account, from, to time.Time) ([]domain.Event, []string, int) {
	var all []domain.Event
	var failed []string
	total := 0
	for _, a := range accounts {
		cals := a.Enabled()
		total += len(cals)
		if len(cals) == 0 {
			continue
		}
		token, err := uc.accessToken(ctx, a)
		if err != nil {
			for _, c := range cals {
				failed = append(failed, fmt.Sprintf("%s (%v)", c.Name, err))
			}
			continue
		}
		for _, c := range cals {
			events, err := uc.api.Events(ctx, a.Provider, token, c.ID, from, to)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s (%v)", c.Name, err))
				continue
			}
			for _, e := range events {
				e.Calendar, e.Color = c.Name, c.Color
				all = append(all, e)
			}
		}
	}
	return all, failed, total
}

// createInAccount writes a new event to a calendar of an account.
func (uc *calendar) createInAccount(ctx context.Context, accountID, calendarID string, e domain.Event) (string, error) {
	a, err := uc.accounts.Get(ctx, accountID)
	if err != nil {
		return "", domain.ErrUnknownCalendar
	}
	var cal *domain.AccountCalendar
	for _, c := range a.Enabled() {
		if c.ID == calendarID && c.Writable {
			cal = &c
			break
		}
	}
	if cal == nil {
		return "", domain.ErrUnknownCalendar
	}
	token, err := uc.accessToken(ctx, a)
	if err != nil {
		return "", err
	}
	if err := uc.api.Create(ctx, a.Provider, token, calendarID, e); err != nil {
		return "", kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return cal.Name, nil
}

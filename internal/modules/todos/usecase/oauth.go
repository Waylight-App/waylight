package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/shared/kernel"
)

// OAuth signs in to providers with the user's own app registration (output
// port, #171); the module wires it to the platform's OAuth client.
type OAuth interface {
	// RedirectURL is the address to register at the provider; "" when
	// PUBLIC_URL is not set (signing in is not available then).
	RedirectURL() string
	NewState() string
	AuthURL(kind domain.SourceKind, clientID, state, verifier string) string
	Exchange(ctx context.Context, kind domain.SourceKind, clientID, clientSecret, code, verifier string) (domain.OAuthGrant, error)
	// Refresh fails with an error wrapping ErrGrantRejected when the user
	// has to sign in again.
	Refresh(ctx context.Context, kind domain.SourceKind, clientID string, g domain.OAuthGrant) (domain.OAuthGrant, error)
	Revoke(ctx context.Context, kind domain.SourceKind, clientID string, g domain.OAuthGrant) error
}

// ErrGrantRejected is wrapped by OAuth.Refresh for refused grants.
var ErrGrantRejected = errors.New("grant rejected")

// TokenSource hands out valid access tokens of signed-in lists, refreshing
// and storing them when they expire (port for the provider adapters).
type TokenSource interface {
	AccessToken(ctx context.Context, s domain.Source) (string, error)
}

type tokenSource struct {
	d  Deps
	mu sync.Mutex
}

// NewTokenSource returns the TokenSource of d (Sources, OAuth and Clock).
func NewTokenSource(d Deps) TokenSource { return &tokenSource{d: d} }

// refreshMargin renews tokens shortly before they expire.
const refreshMargin = 2 * time.Minute

func (ts *tokenSource) AccessToken(ctx context.Context, s domain.Source) (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	// Another request may have refreshed it meanwhile.
	if stored, err := ts.d.Sources.Get(ctx, s.ID); err == nil && stored.Kind == s.Kind {
		s = stored
	}
	g := s.Grant
	if g.Expiry.IsZero() || ts.d.Clock.Now().Add(refreshMargin).Before(g.Expiry) || g.Refresh == "" {
		return g.Access, nil
	}
	ng, err := ts.d.OAuth.Refresh(ctx, s.Kind, s.Username, g)
	if errors.Is(err, ErrGrantRejected) {
		return "", domain.ErrSignInAgain
	}
	if err != nil {
		return "", err
	}
	s.Grant = ng
	if s.ID != "" {
		if err := ts.d.Sources.Save(ctx, s); err != nil {
			return "", fmt.Errorf("store refreshed token: %w", err)
		}
	}
	return ng.Access, nil
}

// pendingSignIn is a sign-in between StartOAuth and Save (in memory).
type pendingSignIn struct {
	kind         domain.SourceKind
	clientID     string
	clientSecret string
	verifier     string
	started      time.Time
	grant        domain.OAuthGrant
	err          error
	done         bool
}

// signInTTL is how long a started sign-in can be finished.
const signInTTL = 15 * time.Minute

type signIns struct {
	mu sync.Mutex
	m  map[string]*pendingSignIn
}

var (
	errSignInUnknown = kernel.NewPublicError(kernel.ErrNotFound, "The sign-in was not started here or has expired. Please start again.")
	errNoPublicURL   = kernel.NewValidationError("kind", "Signing in needs PUBLIC_URL (an https:// address of trisse-home).")
	errSignInFailed  = kernel.NewPublicError(kernel.ErrUnavailable, "The sign-in did not succeed. Check client ID, client secret and the registered redirect address.")
)

// OAuthStart is where the browser signs in.
type OAuthStart struct {
	URL   string
	State string
}

// OAuthInfo says whether signing in is available and what to register.
type OAuthInfo struct {
	RedirectURL string
}

func (uc *sources) OAuthInfo() OAuthInfo {
	if uc.d.OAuth == nil {
		return OAuthInfo{}
	}
	return OAuthInfo{RedirectURL: uc.d.OAuth.RedirectURL()}
}

func (uc *sources) prune(now time.Time) {
	for k, p := range uc.pending.m {
		if now.Sub(p.started) > signInTTL {
			delete(uc.pending.m, k)
		}
	}
}

func (uc *sources) StartOAuth(_ context.Context, in SourceInput) (OAuthStart, error) {
	kind := domain.SourceKind(in.Kind)
	if !kind.OAuth() {
		return OAuthStart{}, kernel.NewValidationError("kind", "Unknown kind of to-do list.")
	}
	if uc.d.OAuth == nil || uc.d.OAuth.RedirectURL() == "" {
		return OAuthStart{}, errNoPublicURL
	}
	if in.Username == "" || len(in.Username) > 200 || len(in.Password) > 500 {
		return OAuthStart{}, kernel.NewValidationError("username", "Enter the client ID of your app registration.")
	}
	state, verifier := uc.d.OAuth.NewState(), uc.d.OAuth.NewState()
	now := uc.d.Clock.Now()
	uc.pending.mu.Lock()
	uc.prune(now)
	if len(uc.pending.m) >= 20 {
		uc.pending.mu.Unlock()
		return OAuthStart{}, kernel.NewValidationError("kind", "Too many sign-ins at once. Please wait a few minutes.")
	}
	uc.pending.m[state] = &pendingSignIn{kind: kind, clientID: in.Username, clientSecret: in.Password, verifier: verifier, started: now}
	uc.pending.mu.Unlock()
	return OAuthStart{URL: uc.d.OAuth.AuthURL(kind, in.Username, state, verifier), State: state}, nil
}

func (uc *sources) FinishOAuth(ctx context.Context, state, code, providerError string) error {
	uc.pending.mu.Lock()
	p := uc.pending.m[state]
	uc.pending.mu.Unlock()
	if p == nil || state == "" {
		return errSignInUnknown
	}
	var g domain.OAuthGrant
	var err error
	if providerError != "" || code == "" {
		err = fmt.Errorf("provider answered %q", providerError)
	} else {
		g, err = uc.d.OAuth.Exchange(ctx, p.kind, p.clientID, p.clientSecret, code, p.verifier)
	}
	uc.pending.mu.Lock()
	defer uc.pending.mu.Unlock()
	p.done, p.grant, p.err = true, g, err
	if err != nil {
		return kernel.TargetFailure(kernel.ErrUnavailable, err)
	}
	return nil
}

// OAuthStatus is "pending", "ready" or "failed".
func (uc *sources) OAuthStatus(state string) (string, error) {
	uc.pending.mu.Lock()
	defer uc.pending.mu.Unlock()
	p := uc.pending.m[state]
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

// signedIn returns a finished sign-in.
func (uc *sources) signedIn(state string) (*pendingSignIn, error) {
	uc.pending.mu.Lock()
	defer uc.pending.mu.Unlock()
	p := uc.pending.m[state]
	switch {
	case p == nil:
		return nil, errSignInUnknown
	case !p.done || p.err != nil:
		return nil, errSignInFailed
	}
	return p, nil
}

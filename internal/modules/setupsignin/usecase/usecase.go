package usecase

import (
	"context"
	"sync"

	"waylight/internal/modules/setupsignin/domain"
	"waylight/internal/shared/kernel"
)

// Started is what the browser needs: the sign-in to follow and the page the
// user signs in on. Neither is a credential.
type Started struct {
	ID       string
	LoginURL string
}

// NextcloudLogin speaks Nextcloud's Login Flow v2 (output port).
type NextcloudLogin interface {
	// Start begins a login flow at base and returns the page the user
	// signs in on and the token to poll with.
	Start(ctx context.Context, base string) (login, pollToken string, err error)
	// Poll asks whether the user has signed in; done stays false until then.
	Poll(ctx context.Context, base, pollToken string) (username, appPassword string, done bool, err error)
}

// HomeAssistantAuth speaks Home Assistant's OAuth (output port). Home
// Assistant follows IndieAuth: the client ID is the client's own address,
// so nothing has to be registered.
type HomeAssistantAuth interface {
	// ClientID is Waylight's address (PUBLIC_URL); "" when it is not set,
	// in which case the sign-in is not available.
	ClientID() string
	// AuthURL is Home Assistant's authorise page for base.
	AuthURL(base, state string) string
	// Exchange trades the code from the redirect for tokens.
	Exchange(ctx context.Context, base, code string) (kernel.SignInGrant, error)
}

// Deps bundles the output ports.
type Deps struct {
	Nextcloud     NextcloudLogin
	HomeAssistant HomeAssistantAuth
	Clock         kernel.Clock
	// Token returns an unguessable random string for IDs and states.
	Token func() string
}

// SignIns is the input port: the assistant starts and follows a sign-in,
// the module that is set up redeems it.
type SignIns interface {
	Start(ctx context.Context, service, address string) (Started, error)
	// Status polls Nextcloud when due and reports pending or ready; a
	// failed sign-in answers with its reason as the error.
	Status(ctx context.Context, id string) (domain.Status, error)
	// Cancel forgets a sign-in, e.g. when the assistant closes.
	Cancel(ctx context.Context, id string) error
	// Callback finishes a Home Assistant sign-in from the redirect. It
	// arrives without a session; the state is the only link.
	Callback(ctx context.Context, state, code, providerError string) error
	kernel.SignIns
}

var (
	errNoPublicURL = kernel.NewPublicError(kernel.ErrInvalid,
		"Signing in with Home Assistant needs PUBLIC_URL, the address you open Waylight with. Create a long-lived access token instead.")
	errTooMany = kernel.NewPublicError(kernel.ErrTooManyRequests, "Too many sign-ins at once. Please wait a few minutes.")
	errDenied  = kernel.NewPublicError(kernel.ErrInvalid, "The sign-in was cancelled. Sign in again or create a key instead.")
	errBadPage = kernel.NewPublicError(kernel.ErrInvalid, "The service did not name a sign-in page. Create a key instead.")
)

type signIns struct {
	d       Deps
	mu      sync.Mutex
	pending map[string]*domain.SignIn
}

// New returns the sign-in use cases.
func New(d Deps) SignIns {
	return &signIns{d: d, pending: map[string]*domain.SignIn{}}
}

// account is who a sign-in belongs to. Guests, kiosk displays and API
// tokens set nothing up.
func account(ctx context.Context) (string, error) {
	p := kernel.PrincipalFrom(ctx)
	if p.TabScoped() || p.IsToken() {
		return "", kernel.ErrForbidden
	}
	return p.SettingsAccount(), nil
}

func (uc *signIns) Start(ctx context.Context, service, address string) (Started, error) {
	acc, err := account(ctx)
	if err != nil {
		return Started{}, err
	}
	s := domain.Service(service)
	if !s.Valid() {
		return Started{}, domain.ErrUnknownService
	}
	base, err := domain.NormaliseURL(address)
	if err != nil {
		return Started{}, err
	}
	if s == domain.HomeAssistant && uc.d.HomeAssistant.ClientID() == "" {
		return Started{}, errNoPublicURL
	}
	if uc.full() {
		return Started{}, errTooMany
	}
	in := domain.New(uc.d.Token(), acc, s, base, uc.d.Clock.Now())
	var login string
	switch s {
	case domain.Nextcloud:
		page, token, err := uc.d.Nextcloud.Start(ctx, base)
		if err != nil {
			return Started{}, kernel.SetupFailure("Nextcloud", err)
		}
		login, in.PollToken = page, token
	case domain.HomeAssistant:
		in.State = uc.d.Token()
		login = uc.d.HomeAssistant.AuthURL(base, in.State)
	}
	if !domain.ValidPage(login) {
		return Started{}, errBadPage
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.sweep()
	if len(uc.pending) >= domain.MaxPending {
		return Started{}, errTooMany
	}
	// One sign-in per account and service: a new start replaces an
	// abandoned one instead of piling up.
	for id, o := range uc.pending {
		if o.Account == acc && o.Service == s {
			delete(uc.pending, id)
		}
	}
	uc.pending[in.ID] = in
	return Started{ID: in.ID, LoginURL: login}, nil
}

func (uc *signIns) full() bool {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.sweep()
	return len(uc.pending) >= domain.MaxPending
}

// sweep drops expired sign-ins; callers hold mu.
func (uc *signIns) sweep() {
	now := uc.d.Clock.Now()
	for id, s := range uc.pending {
		if s.Expired(now) {
			delete(uc.pending, id)
		}
	}
}

// own returns the caller's live sign-in; callers hold mu.
func (uc *signIns) own(acc, id string) *domain.SignIn {
	uc.sweep()
	s := uc.pending[id]
	if s == nil || s.Account != acc {
		return nil
	}
	return s
}

func (uc *signIns) Status(ctx context.Context, id string) (domain.Status, error) {
	acc, err := account(ctx)
	if err != nil {
		return "", err
	}
	uc.mu.Lock()
	s := uc.own(acc, id)
	if s == nil {
		uc.mu.Unlock()
		return "", kernel.ErrSignInUnknown
	}
	now := uc.d.Clock.Now()
	if !s.PollDue(now) {
		status, failure := s.Status, s.Failure
		uc.mu.Unlock()
		return result(status, failure)
	}
	// The poll goes over the network, so the lock is released; LastPoll
	// already claims this round, so concurrent requests do not poll twice.
	s.LastPoll = now
	base, token := s.URL, s.PollToken
	uc.mu.Unlock()

	user, password, done, err := uc.d.Nextcloud.Poll(ctx, base, token)

	uc.mu.Lock()
	defer uc.mu.Unlock()
	if s = uc.own(acc, id); s == nil {
		return "", kernel.ErrSignInUnknown // cancelled meanwhile
	}
	switch {
	case err != nil:
		if s.PollFailed() {
			s.Fail(kernel.SetupFailure("Nextcloud", err))
		}
	case done:
		s.Finish(kernel.SignInGrant{URL: s.URL, Username: user, Secret: password}, uc.d.Clock.Now())
	default:
		s.PollFailures = 0
	}
	return result(s.Status, s.Failure)
}

func result(status domain.Status, failure error) (domain.Status, error) {
	if status == domain.Failed {
		return status, failure
	}
	return status, nil
}

func (uc *signIns) Cancel(ctx context.Context, id string) error {
	acc, err := account(ctx)
	if err != nil {
		return err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	if s := uc.own(acc, id); s != nil {
		delete(uc.pending, id)
	}
	return nil
}

func (uc *signIns) Callback(ctx context.Context, state, code, providerError string) error {
	if state == "" {
		return kernel.ErrSignInUnknown
	}
	uc.mu.Lock()
	uc.sweep()
	var s *domain.SignIn
	for _, p := range uc.pending {
		if p.Service == domain.HomeAssistant && p.Status == domain.Pending && p.State != "" && p.State == state {
			s = p
			break
		}
	}
	if s == nil {
		uc.mu.Unlock()
		return kernel.ErrSignInUnknown
	}
	// Single use: a replayed redirect finds nothing.
	s.State = ""
	base := s.URL
	uc.mu.Unlock()

	var grant kernel.SignInGrant
	var failure error
	switch {
	case providerError != "" || code == "":
		failure = errDenied
	default:
		g, err := uc.d.HomeAssistant.Exchange(ctx, base, code)
		if err != nil {
			failure = kernel.SetupFailure("Home Assistant", err)
		}
		grant = g
		grant.URL = base
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	if uc.pending[s.ID] != s {
		return kernel.ErrSignInUnknown // cancelled or replaced meanwhile
	}
	if failure != nil {
		s.Fail(failure)
		return failure
	}
	s.Finish(grant, uc.d.Clock.Now())
	return nil
}

func (uc *signIns) Redeem(ctx context.Context, id, service string) (kernel.SignInGrant, error) {
	acc, err := account(ctx)
	if err != nil {
		return kernel.SignInGrant{}, err
	}
	uc.mu.Lock()
	defer uc.mu.Unlock()
	s := uc.own(acc, id)
	if s == nil || s.Status != domain.Ready || string(s.Service) != service {
		return kernel.SignInGrant{}, kernel.ErrSignInUnknown
	}
	delete(uc.pending, id)
	return s.Grant, nil
}

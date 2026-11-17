package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/auth/domain"
	"trisse-home/internal/shared/kernel"
)

// ssoStateTTL bounds the time between leaving for the provider and coming back.
const ssoStateTTL = 10 * time.Minute

// maxPendingSSO bounds memory for started but unfinished sign-ins.
const maxPendingSSO = 1000

// OIDCProvider is the OpenID Connect relying party (output port).
type OIDCProvider interface {
	AuthCodeURL(ctx context.Context, state, nonce, verifier string) (string, error)
	// Exchange redeems the code and returns the verified identity.
	Exchange(ctx context.Context, code, verifier, nonce string) (domain.ExternalIdentity, error)
}

// ExternalAccounts decides who besides the owner signs in through OpenID
// Connect (output port, #302). It is set only in builds with the ee/ single
// sign-on; without it, SSO stays the owner's (ADR 0013). Resolve returns the
// account to sign in, creating or updating it as needed, or an error to
// refuse the person.
type ExternalAccounts interface {
	Resolve(ctx context.Context, id domain.ExternalIdentity) (accountID string, err error)
}

// SSOConfig configures single sign-on for the owner.
type SSOConfig struct {
	// OIDC is nil when OpenID Connect is not configured.
	OIDC    OIDCProvider
	Mapping domain.IdentityMapping
	// HeaderUsers may sign in via the trusted proxy header; empty disables it.
	HeaderUsers []string
	// LocalLogin allows the owner's password login (guests always can).
	LocalLogin bool
	// Require2FA asks for the local TOTP code after SSO when it is enabled.
	Require2FA bool
	// Button names the provider on the sign-in page (#514).
	Button SSOButton
	// AutoRedirect sends the sign-in page straight to the provider when no
	// password form is offered (#514).
	AutoRedirect bool
	// Issuer is the provider's issuer URL; linked sign-ins are stored per
	// issuer, so a different provider never matches an old link (#527).
	Issuer string
	// Identities stores linked sign-ins (#527); nil disables linking.
	Identities IdentityRepository
}

// SSOButton is how the sign-in page labels the single sign-on button.
type SSOButton struct {
	// Label is the provider's name ("Microsoft"); empty for the generic text.
	Label string
	// Icon is one of the bundled icons ("microsoft", "key") or empty.
	Icon string
}

// SSOStatus tells the login page which methods exist.
type SSOStatus struct {
	OIDC         bool
	Header       bool
	LocalLogin   bool
	Button       SSOButton
	AutoRedirect bool
}

// SSO signs the owner in through an identity provider or a trusted proxy
// (input port). All methods end in the normal session creation.
type SSO interface {
	Status() SSOStatus
	// BeginOIDC returns the provider URL and the state to bind to the browser.
	BeginOIDC(ctx context.Context) (redirect, state string, err error)
	// CompleteOIDC finishes the code flow. With Require2FA it returns
	// domain.ErrSSOSecondFactorNeed and a ticket for CompleteSecondFactor.
	CompleteOIDC(ctx context.Context, state, code string, client domain.Client) (SessionResponse, string, error)
	// HeaderLogin trusts the user name a trusted proxy authenticated.
	HeaderLogin(ctx context.Context, user string, client domain.Client) (SessionResponse, string, error)
	CompleteSecondFactor(ctx context.Context, ticket, code string, client domain.Client) (SessionResponse, error)
	// BeginLink starts linking a sign-in at the provider to the signed-in
	// account (#527); CompleteOIDC then returns a response with Linked set.
	BeginLink(ctx context.Context, accountID string) (redirect, state string, err error)
	// Links lists the sign-ins linked to an account.
	Links(ctx context.Context, accountID string) ([]domain.IdentityLink, error)
	// Unlink removes one of the account's linked sign-ins.
	Unlink(ctx context.Context, accountID, subject string, client domain.Client) error
	// SetExternalAccounts lets other people sign in through OpenID Connect
	// (#302). It is called once while the server is built.
	SetExternalAccounts(e ExternalAccounts)
}

type pendingSSO struct {
	nonce, verifier string
	expires         time.Time
	// link is the account that started linking a sign-in (#527), "" for a
	// sign-in.
	link string
}

// pendingTicket waits for the local second factor of one account.
type pendingTicket struct {
	account string
	expires time.Time
}

type sso struct {
	deps Deps
	cfg  SSOConfig

	mu       sync.Mutex
	pending  map[string]pendingSSO    // state -> flow data
	tickets  map[string]pendingTicket // second-factor tickets
	external ExternalAccounts
}

// NewSSO returns the single sign-on use cases.
func NewSSO(deps Deps, cfg SSOConfig) SSO {
	return &sso{deps: deps, cfg: cfg, pending: map[string]pendingSSO{}, tickets: map[string]pendingTicket{}}
}

// SetExternalAccounts implements SSO.
func (uc *sso) SetExternalAccounts(e ExternalAccounts) {
	uc.mu.Lock()
	uc.external = e
	uc.mu.Unlock()
}

// Status implements SSO.
func (uc *sso) Status() SSOStatus {
	st := SSOStatus{OIDC: uc.cfg.OIDC != nil, Header: len(uc.cfg.HeaderUsers) > 0, LocalLogin: uc.cfg.LocalLogin}
	if st.OIDC {
		st.Button = uc.cfg.Button
		// Only without a password form: otherwise people could never reach it.
		st.AutoRedirect = uc.cfg.AutoRedirect && !uc.cfg.LocalLogin
	}
	return st
}

func (uc *sso) token() (string, error) {
	t, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return t, nil
}

func (uc *sso) prune(now time.Time) {
	for k, p := range uc.pending {
		if now.After(p.expires) {
			delete(uc.pending, k)
		}
	}
	for k, t := range uc.tickets {
		if now.After(t.expires) {
			delete(uc.tickets, k)
		}
	}
}

// BeginOIDC implements SSO.
func (uc *sso) BeginOIDC(ctx context.Context) (string, string, error) {
	if uc.cfg.OIDC == nil {
		return "", "", domain.ErrSSOUnavailable
	}
	return uc.begin(ctx, "")
}

// begin starts the code flow; link is the account that links the sign-in,
// or "" to sign in.
func (uc *sso) begin(ctx context.Context, link string) (string, string, error) {
	state, err := uc.token()
	if err != nil {
		return "", "", err
	}
	nonce, err := uc.token()
	if err != nil {
		return "", "", err
	}
	verifier, err := uc.token()
	if err != nil {
		return "", "", err
	}
	url, err := uc.cfg.OIDC.AuthCodeURL(ctx, state, nonce, verifier)
	if err != nil {
		return "", "", kernel.NewPublicErrorf(kernel.ErrUnavailable, "The identity provider could not be reached: {detail}", "detail", err.Error())
	}
	now := uc.deps.Clock.Now()
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.prune(now)
	if len(uc.pending) >= maxPendingSSO {
		return "", "", domain.ErrTooManyAttempts
	}
	uc.pending[state] = pendingSSO{nonce: nonce, verifier: verifier, expires: now.Add(ssoStateTTL), link: link}
	return url, state, nil
}

// CompleteOIDC implements SSO.
func (uc *sso) CompleteOIDC(ctx context.Context, state, code string, client domain.Client) (SessionResponse, string, error) {
	if uc.cfg.OIDC == nil {
		return SessionResponse{}, "", domain.ErrSSOUnavailable
	}
	now := uc.deps.Clock.Now()
	uc.mu.Lock()
	p, ok := uc.pending[state]
	delete(uc.pending, state) // single use
	uc.mu.Unlock()
	if !ok || state == "" || now.After(p.expires) {
		return SessionResponse{}, "", domain.ErrSSOStateInvalid
	}
	id, err := uc.cfg.OIDC.Exchange(ctx, code, p.verifier, p.nonce)
	if err != nil {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, uc.deps.Limiter.RecordFailure(client.IP, now))
		return SessionResponse{}, "", kernel.NewPublicErrorf(kernel.ErrUnauthorized, "The identity provider did not confirm the sign-in: {detail}", "detail", err.Error())
	}
	id.Provider = "oidc"
	if p.link != "" {
		return uc.completeLink(ctx, p.link, id, client)
	}
	return uc.finish(ctx, id, client)
}

// HeaderLogin implements SSO; the caller has checked that the request came
// from a trusted proxy.
func (uc *sso) HeaderLogin(ctx context.Context, user string, client domain.Client) (SessionResponse, string, error) {
	if len(uc.cfg.HeaderUsers) == 0 {
		return SessionResponse{}, "", domain.ErrSSOUnavailable
	}
	id := domain.ExternalIdentity{Provider: "header", Subject: user}
	if user == "" || !(domain.IdentityMapping{Subjects: uc.cfg.HeaderUsers}).Allows(id) {
		uc.deps.Events.loginFailed(ctx, domain.OwnerID, client, false)
		return SessionResponse{}, "", domain.ErrSSONotAllowed
	}
	return uc.finish(ctx, id, client)
}

func (uc *sso) finish(ctx context.Context, id domain.ExternalIdentity, client domain.Client) (SessionResponse, string, error) {
	accountID := domain.OwnerID
	// A sign-in linked to an account (#527) wins over the lists of the
	// configuration: the person proved both sides while signed in.
	linked := ""
	if id.Provider == "oidc" {
		var err error
		if linked, err = uc.linkedAccount(ctx, id); err != nil {
			uc.deps.Events.record(ctx, domain.EventLoginFailed, domain.OwnerID, client, "sso: linked account not allowed")
			return SessionResponse{}, "", domain.ErrSSONotAllowed
		}
	}
	if linked != "" {
		accountID = linked
	} else if id.Provider == "oidc" && !uc.cfg.Mapping.Allows(id) {
		uc.mu.Lock()
		external := uc.external
		uc.mu.Unlock()
		resolved, err := uc.resolveExternal(ctx, external, id)
		if err != nil {
			uc.deps.Events.record(ctx, domain.EventLoginFailed, domain.OwnerID, client, "sso: "+id.Subject+" not allowed")
			return SessionResponse{}, "", domain.ErrSSONotAllowed
		}
		accountID = resolved
	}
	account, err := uc.deps.Accounts.Get(ctx, accountID)
	if errors.Is(err, domain.ErrAccountNotFound) && accountID == domain.OwnerID {
		return SessionResponse{}, "", domain.ErrSetupRequired
	}
	if err != nil {
		return SessionResponse{}, "", err
	}
	if account.Disabled() {
		uc.deps.Events.record(ctx, domain.EventLoginFailed, accountID, client, "sso: account disabled")
		return SessionResponse{}, "", domain.ErrSSONotAllowed
	}
	if uc.cfg.Require2FA && account.TOTPEnabled() {
		ticket, err := uc.token()
		if err != nil {
			return SessionResponse{}, "", err
		}
		uc.mu.Lock()
		uc.prune(uc.deps.Clock.Now())
		uc.tickets[ticket] = pendingTicket{account: accountID, expires: uc.deps.Clock.Now().Add(ssoStateTTL)}
		uc.mu.Unlock()
		return SessionResponse{}, ticket, domain.ErrSSOSecondFactorNeed
	}
	uc.deps.Limiter.Reset(client.IP)
	uc.deps.Events.loginSucceeded(ctx, accountID, client, id.Provider)
	res, err := uc.deps.issuer().issue(ctx, accountID, client)
	return res, "", err
}

// resolveExternal asks ExternalAccounts about someone who is not the owner.
// Only members can be signed in this way: the port must not hand out the
// owner's or a guest's account.
func (uc *sso) resolveExternal(ctx context.Context, external ExternalAccounts, id domain.ExternalIdentity) (string, error) {
	if external == nil {
		return "", domain.ErrSSONotAllowed
	}
	accountID, err := external.Resolve(ctx, id)
	if err != nil {
		return "", err
	}
	a, err := uc.deps.Accounts.Get(ctx, accountID)
	if err != nil {
		return "", err
	}
	if !a.IsPerson() {
		return "", domain.ErrSSONotAllowed
	}
	return accountID, nil
}

// CompleteSecondFactor implements SSO.
func (uc *sso) CompleteSecondFactor(ctx context.Context, ticket, code string, client domain.Client) (SessionResponse, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(client.IP, now) {
		return SessionResponse{}, domain.ErrTooManyAttempts
	}
	uc.mu.Lock()
	t, ok := uc.tickets[ticket]
	uc.mu.Unlock()
	if !ok || now.After(t.expires) {
		return SessionResponse{}, domain.ErrSSOStateInvalid
	}
	account, err := uc.deps.Accounts.Get(ctx, t.account)
	if err != nil {
		return SessionResponse{}, err
	}
	if err := uc.deps.checkSecondFactor(ctx, &account, code, now); err != nil {
		uc.deps.Events.loginFailed(ctx, t.account, client, uc.deps.Limiter.RecordFailure(client.IP, now))
		return SessionResponse{}, err
	}
	if err := uc.deps.Accounts.Update(ctx, account); err != nil {
		return SessionResponse{}, fmt.Errorf("update account: %w", err)
	}
	uc.mu.Lock()
	delete(uc.tickets, ticket)
	uc.mu.Unlock()
	uc.deps.Limiter.Reset(client.IP)
	uc.deps.Events.loginSucceeded(ctx, t.account, client, "sso+totp")
	return uc.deps.issuer().issue(ctx, t.account, client)
}

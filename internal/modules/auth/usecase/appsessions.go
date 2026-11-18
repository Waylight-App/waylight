package usecase

import (
	"context"
	"fmt"
	"strings"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/shared/kernel"
)

// AppSessions signs the mobile app in with a session of its own that it
// sends as bearer token (#559, ADR 0024).
type AppSessions interface {
	// Create signs a device in with the credentials of a login or a pairing
	// code. A missing second factor yields *SecondFactorRequired, as for
	// the login.
	Create(ctx context.Context, req AppSessionRequest) (AppSessionResponse, error)
	// NewPairingCode returns a single-use code for the signed-in account,
	// shown as QR code in the web interface.
	NewPairingCode(ctx context.Context) (PairingCodeResponse, error)
	// NewHandoverCode returns the code a browser hands to the app after the
	// person signed in there and confirmed it (#560). Only the app that
	// sent challenge can redeem it.
	NewHandoverCode(ctx context.Context, challenge string) (PairingCodeResponse, error)
}

type appSessions struct {
	deps    Deps
	login   *login
	pairing PairingCodeRepository
}

// NewAppSessions returns the AppSessions use cases.
func NewAppSessions(deps Deps, pairing PairingCodeRepository) AppSessions {
	return &appSessions{deps: deps, login: &login{deps: deps}, pairing: pairing}
}

// Create implements AppSessions.
func (uc *appSessions) Create(ctx context.Context, req AppSessionRequest) (AppSessionResponse, error) {
	device, err := domain.NewDeviceName(req.Device)
	if err != nil {
		return AppSessionResponse{}, err
	}
	if code := strings.TrimSpace(req.PairingCode); code != "" {
		return uc.pair(ctx, code, req.Verifier, req.Login, device)
	}
	account, client, method, err := uc.login.verify(ctx, req.Login)
	if err != nil {
		return AppSessionResponse{}, err
	}
	uc.deps.Events.loginSucceeded(ctx, account.ID(), client, "app ("+method+"): "+device)
	return uc.deps.issuer().issueApp(ctx, account.ID(), client, device)
}

// pair redeems a pairing code. Failures count against the client like
// failed logins, so codes cannot be guessed.
func (uc *appSessions) pair(ctx context.Context, code, verifier string, login LoginRequest, device string) (AppSessionResponse, error) {
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(login.ClientKey, now) {
		return AppSessionResponse{}, domain.ErrTooManyAttempts
	}
	client := domain.NewClient(login.UserAgent, login.ClientKey)
	c, err := uc.pairing.Take(ctx, domain.HashToken(code), now)
	if err == nil && !c.AcceptsVerifier(verifier) {
		err = domain.ErrInvalidPairingCode
	}
	if err != nil {
		uc.deps.Limiter.RecordFailure(login.ClientKey, now)
		return AppSessionResponse{}, err
	}
	// The account may have been disabled since the code was shown.
	account, err := uc.deps.Accounts.Get(ctx, c.AccountID)
	if err != nil || account.Disabled() {
		return AppSessionResponse{}, domain.ErrInvalidPairingCode
	}
	uc.deps.Limiter.Reset(login.ClientKey)
	method := "pairing code"
	if c.Challenge != "" {
		method = "browser"
	}
	uc.deps.Events.loginSucceeded(ctx, account.ID(), client, "app ("+method+"): "+device)
	return uc.deps.issuer().issueApp(ctx, account.ID(), client, device)
}

// NewPairingCode implements AppSessions.
func (uc *appSessions) NewPairingCode(ctx context.Context) (PairingCodeResponse, error) {
	return uc.newCode(ctx, func(hash domain.TokenHash, accountID string) (domain.PairingCode, error) {
		return domain.NewPairingCode(hash, accountID, uc.deps.Clock.Now())
	})
}

// NewHandoverCode implements AppSessions.
func (uc *appSessions) NewHandoverCode(ctx context.Context, challenge string) (PairingCodeResponse, error) {
	return uc.newCode(ctx, func(hash domain.TokenHash, accountID string) (domain.PairingCode, error) {
		return domain.NewHandoverCode(hash, accountID, challenge, uc.deps.Clock.Now())
	})
}

// newCode stores a code built by build for the signed-in account. Kiosk
// displays cannot sign in an app: their link signs in a display, not a
// person.
func (uc *appSessions) newCode(ctx context.Context, build func(domain.TokenHash, string) (domain.PairingCode, error)) (PairingCodeResponse, error) {
	p := kernel.PrincipalFrom(ctx)
	if p.AccountID == "" || p.Role == string(domain.RoleKiosk) {
		return PairingCodeResponse{}, kernel.NewPublicError(kernel.ErrForbidden, "This account cannot pair an app.")
	}
	raw, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return PairingCodeResponse{}, fmt.Errorf("generate pairing code: %w", err)
	}
	c, err := build(domain.HashToken(raw), p.AccountID)
	if err != nil {
		return PairingCodeResponse{}, err
	}
	if err := uc.pairing.Create(ctx, c, uc.deps.Clock.Now()); err != nil {
		return PairingCodeResponse{}, err
	}
	return PairingCodeResponse{Code: raw, ExpiresAt: c.ExpiresAt}, nil
}

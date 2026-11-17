package usecase

import (
	"context"
	"errors"

	"waylight/internal/modules/auth/domain"
)

type getStatus struct {
	accounts AccountRepository
	sessions SessionRepository
	clock    Clock
	// check reports a second factor required by policy but missing (#304);
	// nil when no policy applies.
	check *authenticate
}

// NewGetStatus returns the GetStatus use case.
func NewGetStatus(accounts AccountRepository, sessions SessionRepository, clock Clock) GetStatus {
	return &getStatus{accounts: accounts, sessions: sessions, clock: clock}
}

// NewGetStatusWithPolicies also reports when an admin policy requires a
// second factor the account does not have yet (#304).
func NewGetStatusWithPolicies(accounts AccountRepository, sessions SessionRepository, clock Clock, policies *Policies, passkeys PasskeyRepository) GetStatus {
	return &getStatus{accounts: accounts, sessions: sessions, clock: clock, check: &authenticate{policies: policies, passkeys: passkeys}}
}

// Execute implements GetStatus.
func (uc *getStatus) Execute(ctx context.Context, req StatusRequest) (StatusResponse, error) {
	_, exists, err := loadAccount(ctx, uc.accounts)
	if err != nil {
		return StatusResponse{}, err
	}
	if !exists {
		return StatusResponse{SetupRequired: true}, nil
	}
	s, err := validSession(ctx, uc.sessions, req.SessionToken, uc.clock.Now())
	if errors.Is(err, domain.ErrSessionInvalid) {
		return StatusResponse{}, nil
	}
	if err != nil {
		return StatusResponse{}, err
	}
	a, err := uc.accounts.Get(ctx, s.AccountID())
	if err != nil || a.Disabled() {
		return StatusResponse{}, nil //nolint:nilerr // a session of a removed or disabled account is simply not signed in
	}
	res := StatusResponse{Authenticated: true, CSRFToken: s.CSRFToken(), Role: string(a.Role()), Username: a.Name()}
	if uc.check != nil && uc.check.policies.Current(ctx).RequireSecondFactor && a.Role() != domain.RoleKiosk {
		if res.SecondFactorRequired, err = uc.check.secondFactorMissing(ctx, a); err != nil {
			return StatusResponse{}, err
		}
	}
	return res, nil
}

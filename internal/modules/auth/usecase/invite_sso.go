package usecase

import (
	"context"
	"fmt"

	"trisse-home/internal/modules/auth/domain"
)

// InviteSignUp creates an invited account through the provider's sign-in
// instead of a password (#528, output port of the SSO use case; implemented
// by the account links).
type InviteSignUp interface {
	// CheckInvite tells whether token is a usable invitation, before the
	// person is sent to the provider.
	CheckInvite(ctx context.Context, token, clientKey string) error
	// CreateFromInvite creates the invited account with a password nobody
	// knows and consumes the invitation.
	CreateFromInvite(ctx context.Context, token, clientKey string, client domain.Client) (domain.Account, error)
}

// CheckInvite implements InviteSignUp.
func (uc *accountLinks) CheckInvite(ctx context.Context, token, clientKey string) error {
	l, err := uc.find(ctx, token, clientKey)
	if err != nil {
		return err
	}
	if l.Purpose != domain.LinkInvite {
		return domain.ErrInvalidAccountLink
	}
	return uc.withinLimits(ctx, l.Role)
}

// CreateFromInvite implements InviteSignUp.
func (uc *accountLinks) CreateFromInvite(ctx context.Context, token, clientKey string, client domain.Client) (domain.Account, error) {
	l, err := uc.find(ctx, token, clientKey)
	if err != nil {
		return domain.Account{}, err
	}
	if l.Purpose != domain.LinkInvite {
		return domain.Account{}, domain.ErrInvalidAccountLink
	}
	// A password nobody knows, as for members the provider creates (#302):
	// the account signs in through the provider until a password is set.
	secret, err := uc.deps.Tokens.NewToken()
	if err != nil {
		return domain.Account{}, fmt.Errorf("generate password: %w", err)
	}
	hash, err := uc.deps.Hasher.Hash(secret)
	if err != nil {
		return domain.Account{}, fmt.Errorf("hash password: %w", err)
	}
	a, err := uc.inviteAccount(ctx, l, hash)
	if err != nil {
		return domain.Account{}, err
	}
	uc.deps.Events.record(ctx, domain.EventInviteAccepted, a.ID(), client, fmt.Sprintf("%s %s via single sign-on", a.Role(), a.Name()))
	uc.deps.Limiter.Reset(clientKey)
	return a, nil
}

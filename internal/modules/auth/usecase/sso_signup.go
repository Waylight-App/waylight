package usecase

import (
	"context"

	"trisse-home/internal/modules/auth/domain"
)

// SetInviteSignUp implements SSO.
func (uc *sso) SetInviteSignUp(s InviteSignUp) {
	uc.mu.Lock()
	uc.invites = s
	uc.mu.Unlock()
}

func (uc *sso) inviteSignUp() InviteSignUp {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.invites
}

// CanSignUp implements SSO: invited people are members or guests, so the
// licence decides, as for linking a sign-in (#527).
func (uc *sso) CanSignUp(ctx context.Context) bool {
	if uc.cfg.OIDC == nil || uc.deps.Identities == nil || uc.inviteSignUp() == nil {
		return false
	}
	_, link := uc.licensed(ctx)
	return link
}

// BeginSignUp implements SSO.
func (uc *sso) BeginSignUp(ctx context.Context, token, clientKey string) (string, string, error) {
	if !uc.CanSignUp(ctx) {
		return "", "", domain.ErrSSOUnavailable
	}
	if err := uc.inviteSignUp().CheckInvite(ctx, token, clientKey); err != nil {
		return "", "", err
	}
	return uc.beginWith(ctx, pendingSSO{invite: token, inviteClient: clientKey})
}

// completeSignUp creates the invited account for the identity the provider
// confirmed, links it as the account's only way in and signs in.
func (uc *sso) completeSignUp(ctx context.Context, token, clientKey string, id domain.ExternalIdentity, client domain.Client) (SessionResponse, error) {
	if !uc.CanSignUp(ctx) || id.Subject == "" {
		return SessionResponse{}, domain.ErrSSONotAllowed
	}
	// Someone who has an account here already signs in instead.
	if _, err := uc.deps.Identities.Find(ctx, uc.cfg.Issuer, id.Subject); err == nil {
		return SessionResponse{}, domain.ErrSignUpIdentityTaken
	}
	a, err := uc.inviteSignUp().CreateFromInvite(ctx, token, clientKey, client)
	if err != nil {
		return SessionResponse{}, err
	}
	link := domain.IdentityLink{AccountID: a.ID(), Issuer: uc.cfg.Issuer, Subject: id.Subject, Email: id.Email,
		CreatedAt: uc.deps.Clock.Now(), OnlySignIn: true}
	if err := uc.deps.Identities.Add(ctx, link); err != nil {
		return SessionResponse{}, err
	}
	uc.deps.Events.record(ctx, domain.EventIdentityLinked, a.ID(), client, id.Email)
	uc.deps.Events.loginSucceeded(ctx, a.ID(), client, id.Provider)
	return uc.deps.issuer().issue(ctx, a.ID(), client)
}

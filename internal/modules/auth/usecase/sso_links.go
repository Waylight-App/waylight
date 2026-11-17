package usecase

import (
	"context"
	"errors"

	"trisse-home/internal/modules/auth/domain"
)

// IdentityRepository stores sign-ins at the OpenID Connect provider linked to
// accounts (#527, output port).
type IdentityRepository interface {
	// Add stores a link; domain.ErrIdentityTaken when the identity is linked already.
	Add(ctx context.Context, l domain.IdentityLink) error
	// Find returns the account an identity is linked to, or domain.ErrIdentityNotFound.
	Find(ctx context.Context, issuer, subject string) (string, error)
	ListByAccount(ctx context.Context, accountID string) ([]domain.IdentityLink, error)
	Remove(ctx context.Context, accountID, issuer, subject string) error
	// RemoveAccount drops every link of a deleted account.
	RemoveAccount(ctx context.Context, accountID string) error
	// PasswordSet clears OnlySignIn once the account has a known password.
	PasswordSet(ctx context.Context, accountID string) error
}

// passwordSet tells the linked sign-ins that the account has a known
// password now (#528); called wherever a password is set.
func (d Deps) passwordSet(ctx context.Context, accountID string) {
	if d.Identities != nil {
		_ = d.Identities.PasswordSet(ctx, accountID) // the account works either way
	}
}

// SSOLicence is implemented by ExternalAccounts to tell how far single
// sign-on may be used by accounts other than the owner's (#302, #527).
// Without it only the owner links and uses a sign-in.
type SSOLicence interface {
	// SSOLicensed reports whether linked sign-ins may be used, and whether
	// new ones may be linked; an expired licence allows the first only.
	SSOLicensed(ctx context.Context) (signIn, link bool)
}

// licensed asks the licence for an account other than the owner's.
func (uc *sso) licensed(ctx context.Context) (signIn, link bool) {
	uc.mu.Lock()
	external := uc.external
	uc.mu.Unlock()
	lic, ok := external.(SSOLicence)
	if !ok {
		return false, false
	}
	return lic.SSOLicensed(ctx)
}

// mayLink reports whether the account may link a new sign-in: the owner
// always, everyone else with a licence that allows changes.
func (uc *sso) mayLink(ctx context.Context, accountID string) bool {
	if accountID == domain.OwnerID {
		return true
	}
	_, link := uc.licensed(ctx)
	return link
}

// maySignIn reports whether a linked sign-in of the account counts.
func (uc *sso) maySignIn(ctx context.Context, accountID string) bool {
	if accountID == domain.OwnerID {
		return true
	}
	signIn, _ := uc.licensed(ctx)
	return signIn
}

// BeginLink starts the provider's sign-in to link it to the signed-in
// account; the callback stores the link instead of signing in.
func (uc *sso) BeginLink(ctx context.Context, accountID string) (string, string, error) {
	if uc.cfg.OIDC == nil || uc.deps.Identities == nil {
		return "", "", domain.ErrSSOUnavailable
	}
	if !uc.mayLink(ctx, accountID) {
		return "", "", domain.ErrLinkNeedsLicence
	}
	return uc.begin(ctx, accountID)
}

// Links lists the sign-ins linked to an account.
func (uc *sso) Links(ctx context.Context, accountID string) ([]domain.IdentityLink, error) {
	if uc.deps.Identities == nil {
		return nil, nil
	}
	return uc.deps.Identities.ListByAccount(ctx, accountID)
}

// Unlink removes a linked sign-in of the account.
func (uc *sso) Unlink(ctx context.Context, accountID, subject string, client domain.Client) error {
	if uc.deps.Identities == nil {
		return domain.ErrIdentityNotFound
	}
	links, err := uc.deps.Identities.ListByAccount(ctx, accountID)
	if err != nil {
		return err
	}
	for _, l := range links {
		if l.Subject == subject && l.Issuer == uc.cfg.Issuer && l.OnlySignIn {
			return domain.ErrIdentityOnlySignIn
		}
	}
	if err := uc.deps.Identities.Remove(ctx, accountID, uc.cfg.Issuer, subject); err != nil {
		return err
	}
	uc.deps.Events.record(ctx, domain.EventIdentityUnlinked, accountID, client, "")
	return nil
}

// completeLink stores the identity of a finished provider sign-in for the
// account that started it.
func (uc *sso) completeLink(ctx context.Context, accountID string, id domain.ExternalIdentity, client domain.Client) (SessionResponse, string, error) {
	if !uc.mayLink(ctx, accountID) {
		return SessionResponse{}, "", domain.ErrLinkNeedsLicence
	}
	if id.Subject == "" {
		return SessionResponse{}, "", domain.ErrSSONotAllowed
	}
	link := domain.IdentityLink{AccountID: accountID, Issuer: uc.cfg.Issuer, Subject: id.Subject, Email: id.Email, CreatedAt: uc.deps.Clock.Now()}
	if err := uc.deps.Identities.Add(ctx, link); err != nil {
		return SessionResponse{}, "", err
	}
	uc.deps.Events.record(ctx, domain.EventIdentityLinked, accountID, client, id.Email)
	return SessionResponse{Linked: true}, "", nil
}

// linkedAccount returns the account an identity is linked to, "" if none.
// A link of someone other than the owner only counts while licensed.
func (uc *sso) linkedAccount(ctx context.Context, id domain.ExternalIdentity) (string, error) {
	if uc.deps.Identities == nil || id.Subject == "" {
		return "", nil
	}
	accountID, err := uc.deps.Identities.Find(ctx, uc.cfg.Issuer, id.Subject)
	if errors.Is(err, domain.ErrIdentityNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !uc.maySignIn(ctx, accountID) {
		return "", domain.ErrSSONotAllowed
	}
	return accountID, nil
}

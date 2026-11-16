// Package usecase signs members in through the identity provider (#302).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"errors"
	"fmt"

	licence "trisse-home/ee/licence/domain"
	"trisse-home/ee/sso/domain"
	authdomain "trisse-home/internal/modules/auth/domain"
)

// Link connects a provider subject with the member it signs in as.
type Link struct {
	Subject   string
	AccountID string
	// Groups were synced from the provider last time; only these are ever
	// removed again, so groups an admin added by hand stay.
	Groups []string
	// DisabledBySSO marks members this module locked out, so it lets them in
	// again when the provider does, but never undoes an admin's decision.
	DisabledBySSO bool
}

// Links stores the links (output port).
type Links interface {
	// Get returns ok = false for an unknown subject.
	Get(ctx context.Context, subject string) (Link, bool, error)
	Save(ctx context.Context, l Link) error
	Delete(ctx context.Context, subject string) error
}

// Accounts manages members in the auth module (output port, implemented by
// auth.Provisioning).
type Accounts interface {
	CreateMember(ctx context.Context, name string) (string, error)
	UpdateGroups(ctx context.Context, accountID string, join, leave []string) error
	SetDisabled(ctx context.Context, accountID string, disabled bool) error
}

// Licensing tells how far single sign-on may be used (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

// Errors that refuse a sign-in; the auth module answers them all alike, so
// the person learns nothing about the reason.
var (
	ErrNotLicensed = errors.New("single sign-on for members needs a valid licence key")
	ErrNotAllowed  = errors.New("the identity provider's person is not allowed in")
)

// Deps are the ports of the use case.
type Deps struct {
	Links     Links
	Accounts  Accounts
	Licensing Licensing
	Policy    domain.Policy
}

// Resolver decides which member a provider identity signs in as.
type Resolver struct{ d Deps }

// New returns the resolver.
func New(d Deps) *Resolver { return &Resolver{d: d} }

// Resolve creates, updates or locks out the member of an identity. With a
// read-only licence (expired, or more accounts than seats) existing members
// still sign in, but no account is created and nothing is changed.
func (r *Resolver) Resolve(ctx context.Context, id domain.Identity) (string, error) {
	level := r.d.Licensing.Level(ctx, licence.FeatureSSO)
	if level == licence.LevelNone {
		return "", ErrNotLicensed
	}
	allowed := r.d.Policy.Allows(id)
	link, found, err := r.d.Links.Get(ctx, id.Subject)
	if err != nil {
		return "", fmt.Errorf("load link: %w", err)
	}
	if found {
		accountID, err := r.existing(ctx, link, id, allowed, level)
		if !errors.Is(err, authdomain.ErrGuestNotFound) {
			return accountID, err
		}
		// An admin deleted the member: forget the link and start over.
		if err := r.d.Links.Delete(ctx, link.Subject); err != nil {
			return "", fmt.Errorf("delete link: %w", err)
		}
	}
	if !allowed || level != licence.LevelFull {
		return "", ErrNotAllowed
	}
	accountID, err := r.d.Accounts.CreateMember(ctx, id.AccountName())
	if err != nil {
		return "", fmt.Errorf("create member: %w", err)
	}
	link = Link{Subject: id.Subject, AccountID: accountID}
	if err := r.sync(ctx, &link, id); err != nil {
		return "", err
	}
	return accountID, nil
}

func (r *Resolver) existing(ctx context.Context, link Link, id domain.Identity, allowed bool, level licence.Level) (string, error) {
	if !allowed {
		// The provider no longer lets the person in: lock them out, which also
		// ends their open sessions.
		if level == licence.LevelFull && !link.DisabledBySSO {
			if err := r.d.Accounts.SetDisabled(ctx, link.AccountID, true); err != nil {
				return "", err
			}
			link.DisabledBySSO = true
			if err := r.d.Links.Save(ctx, link); err != nil {
				return "", fmt.Errorf("save link: %w", err)
			}
		}
		return "", ErrNotAllowed
	}
	if level != licence.LevelFull {
		return link.AccountID, nil
	}
	if link.DisabledBySSO {
		if err := r.d.Accounts.SetDisabled(ctx, link.AccountID, false); err != nil {
			return "", err
		}
		link.DisabledBySSO = false
	}
	if err := r.sync(ctx, &link, id); err != nil {
		return "", err
	}
	return link.AccountID, nil
}

// sync brings the member's groups in line with the provider and saves the link.
func (r *Resolver) sync(ctx context.Context, link *Link, id domain.Identity) error {
	now := r.d.Policy.GroupsToSync(id)
	if err := r.d.Accounts.UpdateGroups(ctx, link.AccountID, now, domain.Leave(link.Groups, now)); err != nil {
		return err
	}
	link.Groups = now
	if err := r.d.Links.Save(ctx, *link); err != nil {
		return fmt.Errorf("save link: %w", err)
	}
	return nil
}

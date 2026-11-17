package domain

import (
	"time"

	"trisse-home/internal/shared/kernel"
)

// IdentityLink ties an account at the OpenID Connect provider, such as a
// personal Microsoft account, to an account here (#527). Once linked, a
// sign-in with that identity signs in the account, without the e-mail and
// subject lists of the configuration: the person proved both sides while
// signed in.
type IdentityLink struct {
	AccountID string
	// Issuer and Subject identify the person at the provider; the subject is
	// stable and unique per issuer, the e-mail address is neither.
	Issuer  string
	Subject string
	// Email is shown in the settings so people recognise the link; it is
	// never used to decide anything.
	Email     string
	CreatedAt time.Time
	// OnlySignIn: the account was created through this sign-in and has no
	// known password (#528); it cannot be disconnected until one is set.
	OnlySignIn bool
}

var (
	// ErrIdentityTaken is returned when the identity is linked to another account.
	ErrIdentityTaken = kernel.NewPublicError(kernel.ErrConflict, "This sign-in is already linked to another account.")
	// ErrIdentityNotFound is returned for an unknown link.
	ErrIdentityNotFound = kernel.NewPublicError(kernel.ErrNotFound, "This sign-in is not linked.")
	// ErrIdentityOnlySignIn is returned when disconnecting the only way into an account.
	ErrIdentityOnlySignIn = kernel.NewPublicError(kernel.ErrConflict, "This sign-in is the only way into the account. Set a password first, for example with a reset link.")
	// ErrSignUpIdentityTaken is returned when someone accepts an invitation
	// with a sign-in that already belongs to an account here.
	ErrSignUpIdentityTaken = kernel.NewPublicError(kernel.ErrConflict, "This sign-in already belongs to an account here. Sign in with it instead.")
	// ErrDeviceNotLinked is returned when a Microsoft account signs in that
	// is linked to no account here (#535).
	ErrDeviceNotLinked = kernel.NewPublicError(kernel.ErrUnauthorized, "This Microsoft account is not linked to an account here. Sign in with your password and connect it under Settings → Account.")
	// ErrDeviceExpired is returned when the code ran out or was declined.
	ErrDeviceExpired = kernel.NewPublicError(kernel.ErrConflict, "The code is no longer valid. Start again.")
	// ErrLinkNeedsLicence is returned when someone other than the owner links
	// a sign-in without a licence key that includes single sign-on.
	ErrLinkNeedsLicence = kernel.NewPublicError(kernel.ErrForbidden, "Single sign-on for this account needs a licence key with sso.")
)

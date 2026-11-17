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
}

var (
	// ErrIdentityTaken is returned when the identity is linked to another account.
	ErrIdentityTaken = kernel.NewPublicError(kernel.ErrConflict, "This sign-in is already linked to another account.")
	// ErrIdentityNotFound is returned for an unknown link.
	ErrIdentityNotFound = kernel.NewPublicError(kernel.ErrNotFound, "This sign-in is not linked.")
	// ErrLinkNeedsLicence is returned when someone other than the owner links
	// a sign-in without a licence key that includes single sign-on.
	ErrLinkNeedsLicence = kernel.NewPublicError(kernel.ErrForbidden, "Single sign-on for this account needs a licence key with sso.")
)

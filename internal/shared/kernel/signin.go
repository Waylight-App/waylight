package kernel

import (
	"context"
	"time"
)

// Services the setup assistant can sign in to instead of asking for a
// pasted key (#654).
const (
	SignInNextcloud     = "nextcloud"
	SignInHomeAssistant = "homeassistant"
)

// SignInGrant is the credential a finished sign-in yields. Modules store it
// like a pasted key; it never reaches the browser.
type SignInGrant struct {
	// URL is the address the sign-in started from, without a trailing slash.
	URL string
	// Username is the Nextcloud login name; empty for Home Assistant.
	Username string
	// Secret is the Nextcloud app password or the Home Assistant refresh
	// token.
	Secret string
	// Access is a short-lived access token (Home Assistant) and Expiry its
	// end; both are empty for Nextcloud.
	Access string
	Expiry time.Time
	// ClientID is the OAuth client the refresh token was issued to.
	ClientID string
}

// SignIns hands finished sign-ins to the module that stores them.
type SignIns interface {
	// Redeem returns the grant of a finished sign-in of the given service
	// once, and only to the account that started it; afterwards the
	// sign-in is gone. Unknown, unfinished or expired sign-ins fail with
	// ErrSignInUnknown.
	Redeem(ctx context.Context, id, service string) (SignInGrant, error)
}

// ErrSignInUnknown is returned for a sign-in that does not exist (any more)
// for the asking account. It is an input error: the browser session is fine.
var ErrSignInUnknown = NewPublicError(ErrInvalid, "The sign-in has expired or was not finished. Please sign in again.")

// RedeemSignIn redeems a sign-in at s. Without s (the sign-in module is
// switched off) no sign-in exists.
func RedeemSignIn(ctx context.Context, s SignIns, id, service string) (SignInGrant, error) {
	if s == nil {
		return SignInGrant{}, ErrSignInUnknown
	}
	return s.Redeem(ctx, id, service)
}

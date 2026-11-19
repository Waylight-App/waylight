package usecase

import (
	"time"

	"waylight/internal/shared/kernel"
)

// StatusRequest carries the (possibly empty) session token of the caller.
type StatusRequest struct {
	SessionToken string
}

// StatusResponse describes the authentication state of the caller.
type StatusResponse struct {
	SetupRequired bool
	Authenticated bool
	CSRFToken     string
	// Role and Username of the signed-in account.
	Role     string
	Username string
	// SecondFactorRequired: an admin policy requires a second factor and the
	// account has none yet (#304).
	SecondFactorRequired bool
}

// SetupRequest is the input of SetupAccount.
type SetupRequest struct {
	SetupToken string
	Password   string
	ClientKey  string
	UserAgent  string
	// Profile is what the start page shows (#281), passed on unchanged to
	// the SetupObserver; auth does not interpret it.
	Profile string
}

// LoginRequest is the input of Login. ClientKey identifies the client for
// rate limiting (usually its IP address).
type LoginRequest struct {
	// Username is empty for the owner and the guest's name otherwise.
	Username  string
	Password  string
	TOTPCode  string
	ClientKey string
	UserAgent string
}

// TwoFactorStatus describes the second factor.
type TwoFactorStatus struct {
	Enabled           bool
	RecoveryCodesLeft int
}

// TwoFactorEnrollment is shown while setting up the authenticator app.
type TwoFactorEnrollment struct {
	Secret string
	URI    string
}

// DisableTwoFactorRequest requires the password and a current code.
type DisableTwoFactorRequest struct {
	Password  string
	Code      string
	ClientKey string
}

// LogoutRequest is the input of Logout.
type LogoutRequest struct {
	SessionToken string
}

// AuthenticateRequest is the input of Authenticate.
type AuthenticateRequest struct {
	SessionToken string
}

// AuthenticateResponse describes a valid session. Refreshed is true when the
// expiry was extended and the cookie should be re-issued.
type AuthenticateResponse struct {
	CSRFToken string
	ExpiresAt time.Time
	Refreshed bool
	Principal kernel.Principal
}

// ChangePasswordRequest is the input of ChangePassword.
type ChangePasswordRequest struct {
	CurrentPassword string
	NewPassword     string
	ClientKey       string
	UserAgent       string
}

// SessionView describes an active session in the session overview.
type SessionView struct {
	ID         string
	Device     string // empty for browser sessions
	UserAgent  string
	ClientIP   string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	Current    bool
}

// AppSessionRequest signs a device in (#559): either with the credentials
// of a login or with a pairing code.
type AppSessionRequest struct {
	Login       LoginRequest
	PairingCode string
	Device      string
}

// AppSessionResponse carries the bearer token of a new app session; it is
// shown only once.
type AppSessionResponse struct {
	Token     string
	Device    string
	ExpiresAt time.Time
}

// PairingCodeResponse is a pairing code for the QR code of the web interface.
type PairingCodeResponse struct {
	Code      string
	ExpiresAt time.Time
}

// SessionResponse describes a newly created session. SessionToken is the raw
// token for the cookie; it is never stored.
type SessionResponse struct {
	SessionToken string
	CSRFToken    string
	ExpiresAt    time.Time
	// Linked is set instead of a session when the provider's sign-in was
	// linked to an account (#527).
	Linked bool
}

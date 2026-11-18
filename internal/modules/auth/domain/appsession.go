package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// AppSessionTokenPrefix marks the bearer tokens of app sessions, so the
// guard can tell them from personal API tokens (APITokenPrefix) without a
// database lookup.
const AppSessionTokenPrefix = "wa_"

// AppSessionTTL is the sliding lifetime of an app session. A phone is used
// less regularly than a browser tab, so it stays signed in longer; the
// admin's maximum session age still applies (ADR 0024).
const AppSessionTTL = 90 * 24 * time.Hour

// PairingCodeTTL is how long a pairing code shown as QR code is valid.
const PairingCodeTTL = 5 * time.Minute

// maxDeviceNameLength bounds the device name shown in the session overview.
const maxDeviceNameLength = 80

// IsAppSessionToken reports whether raw is the bearer token of an app session.
func IsAppSessionToken(raw string) bool { return strings.HasPrefix(raw, AppSessionTokenPrefix) }

// NewDeviceName validates the name an app sends for its device.
func NewDeviceName(raw string) (string, error) {
	name := strings.Join(strings.Fields(raw), " ")
	if name == "" || utf8.RuneCountInString(name) > maxDeviceNameLength {
		return "", kernel.NewValidationError("device", "must be 1–80 characters")
	}
	return name, nil
}

// NewAppSession starts a session for a device that sends its token as
// bearer token instead of a cookie. It has a CSRF token like every session,
// but the guard does not ask for it: a browser never sends the bearer token
// on its own, so cross-site requests cannot use it.
func NewAppSession(accountID string, tokenHash TokenHash, publicID, csrfToken string, client Client, device string, now time.Time) (Session, error) {
	device, err := NewDeviceName(device)
	if err != nil {
		return Session{}, err
	}
	s, err := NewSession(accountID, tokenHash, publicID, csrfToken, client, now, AppSessionTTL)
	if err != nil {
		return Session{}, err
	}
	s.device = device
	return s, nil
}

// PairingCode lets a signed-in account hand a session to the app: the web
// interface shows it as QR code, the app redeems it once (#559). Only its
// hash is stored.
type PairingCode struct {
	Hash      TokenHash
	AccountID string
	ExpiresAt time.Time
}

// NewPairingCode creates a pairing code that expires after PairingCodeTTL.
func NewPairingCode(hash TokenHash, accountID string, now time.Time) (PairingCode, error) {
	if hash == "" || accountID == "" {
		return PairingCode{}, kernel.NewValidationError("pairing code", "hash and account are required")
	}
	return PairingCode{Hash: hash, AccountID: accountID, ExpiresAt: now.Add(PairingCodeTTL)}, nil
}

// ErrInvalidPairingCode is returned for unknown, used or expired codes.
var ErrInvalidPairingCode = kernel.NewPublicError(kernel.ErrUnauthorized, "This pairing code is not valid (any more). Show a new one in the settings.")

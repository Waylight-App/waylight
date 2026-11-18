package domain

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"regexp"
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

// HandoverCodeTTL is how long the code of a browser handover is valid: the
// browser hands it to the app at once (#560).
const HandoverCodeTTL = time.Minute

// pkceChallenge is an S256 code challenge (RFC 7636): the unpadded
// base64url SHA-256 of the verifier.
var pkceChallenge = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

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
	// Challenge is the PKCE challenge of a browser handover (#560); empty
	// for codes shown as QR code.
	Challenge string
}

// NewPairingCode creates a pairing code that expires after PairingCodeTTL.
func NewPairingCode(hash TokenHash, accountID string, now time.Time) (PairingCode, error) {
	if hash == "" || accountID == "" {
		return PairingCode{}, kernel.NewValidationError("pairing code", "hash and account are required")
	}
	return PairingCode{Hash: hash, AccountID: accountID, ExpiresAt: now.Add(PairingCodeTTL)}, nil
}

// NewHandoverCode creates the code of a browser handover (#560). Only the
// app that sent challenge can redeem it, with the matching verifier, so a
// code read from the redirect by another app is useless.
func NewHandoverCode(hash TokenHash, accountID, challenge string, now time.Time) (PairingCode, error) {
	if !pkceChallenge.MatchString(challenge) {
		return PairingCode{}, kernel.NewValidationError("challenge", "must be an S256 PKCE code challenge")
	}
	c, err := NewPairingCode(hash, accountID, now)
	if err != nil {
		return PairingCode{}, err
	}
	c.Challenge, c.ExpiresAt = challenge, now.Add(HandoverCodeTTL)
	return c, nil
}

// AcceptsVerifier reports whether verifier redeems the code: codes shown as
// QR code take none, handover codes the one that matches their challenge.
func (c PairingCode) AcceptsVerifier(verifier string) bool {
	if c.Challenge == "" {
		return verifier == ""
	}
	sum := sha256.Sum256([]byte(verifier))
	got := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(c.Challenge)) == 1
}

// ErrInvalidPairingCode is returned for unknown, used or expired codes.
var ErrInvalidPairingCode = kernel.NewPublicError(kernel.ErrUnauthorized, "This pairing code is not valid (any more). Show a new one in the settings.")

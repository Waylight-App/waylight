package domain

import (
	"net/mail"
	"strings"
	"time"

	"trisse-home/internal/shared/kernel"
)

// Account links (#296): an admin invites someone with a link, or sends a
// member or guest a link to set a new password, so the admin never has to
// know or hand over a password. Like kiosk links, the token is shown once
// and only its hash is stored; a link works once and expires.

// LinkPurpose says what an account link does.
type LinkPurpose string

// Link purposes. The values are stored.
const (
	// LinkInvite creates a new account with the name and role chosen by
	// the admin; the invited person sets the password.
	LinkInvite LinkPurpose = "invite"
	// LinkReset sets a new password for an existing member or guest.
	LinkReset LinkPurpose = "reset"
)

// Validity of account links.
const (
	// ResetLinkTTL is short: a reset link replaces a password.
	ResetLinkTTL = 24 * time.Hour
	// EmailResetLinkTTL is shorter still: the link travels by e-mail and
	// was asked for by whoever knew the account name.
	EmailResetLinkTTL = time.Hour
	// ResetRequestInterval is the minimum time between two reset mails to
	// the same account, so nobody can flood a mailbox.
	ResetRequestInterval = 5 * time.Minute
	// maxEmailLength is the longest address SMTP allows (RFC 5321).
	maxEmailLength = 254
	// MinInviteDays and MaxInviteDays bound how long an invitation waits.
	MinInviteDays = 1
	MaxInviteDays = 30
	// MaxOpenInvites bounds pending invitations.
	MaxOpenInvites = 20
)

// AccountLink is a pending invitation or password reset.
type AccountLink struct {
	ID        string
	TokenHash TokenHash
	Purpose   LinkPurpose
	// AccountID is the account a reset link is for; empty for invitations.
	AccountID string
	// Name, Role and Tabs describe the account an invitation creates.
	Name      string
	Role      Role
	Tabs      []string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// NewInvite validates an invitation for a member or a guest. The name
// follows the rules of guest and member names.
func NewInvite(id string, hash TokenHash, name string, role Role, tabs []string, days int, now time.Time) (AccountLink, error) {
	if role != RoleMember && role != RoleGuest {
		return AccountLink{}, kernel.NewValidationError("role", "Invite a member or a guest.")
	}
	if !guestNamePattern.MatchString(name) || name == OwnerID {
		return AccountLink{}, kernel.NewValidationError("name", "The name must be 2–32 lowercase letters, digits, dots, dashes or underscores.")
	}
	if days < MinInviteDays || days > MaxInviteDays {
		return AccountLink{}, kernel.NewValidationError("validDays", "An invitation is valid for 1 to 30 days.")
	}
	if id == "" || hash == "" {
		return AccountLink{}, kernel.NewValidationError("invite", "id and token are required")
	}
	if role == RoleMember {
		tabs = nil // members have their own tabs
	}
	return AccountLink{
		ID: id, TokenHash: hash, Purpose: LinkInvite, Name: name, Role: role, Tabs: normaliseTabs(tabs),
		CreatedAt: now, ExpiresAt: now.AddDate(0, 0, days),
	}, nil
}

// NewResetLink creates a password reset link for a member or a guest. The
// owner (today the only admin) keeps the shell-only reset on purpose: a
// link that can take over the admin account must not exist.
func NewResetLink(id string, hash TokenHash, a Account, now time.Time, ttl time.Duration) (AccountLink, error) {
	switch a.Role() {
	case RoleMember, RoleGuest:
	case RoleOwner:
		return AccountLink{}, ErrAdminResetByShell
	default:
		return AccountLink{}, ErrNoPassword // kiosks sign in with their own link
	}
	if id == "" || hash == "" {
		return AccountLink{}, kernel.NewValidationError("reset", "id and token are required")
	}
	return AccountLink{ID: id, TokenHash: hash, Purpose: LinkReset, AccountID: a.ID(), Name: a.Name(), Role: a.Role(), CreatedAt: now, ExpiresAt: now.Add(ttl)}, nil
}

// Expired reports whether the link can no longer be used.
func (l AccountLink) Expired(now time.Time) bool { return !now.Before(l.ExpiresAt) }

// ErrInvalidAccountLink is returned for an unknown, used or expired link.
// It does not say which, so the answer does not help guessing tokens.
var ErrInvalidAccountLink = kernel.NewPublicError(kernel.ErrUnauthorized, "This link is not valid (any more). Ask for a new one.")

// ErrAdminResetByShell is returned when a reset link is asked for the admin.
var ErrAdminResetByShell = kernel.NewPublicError(kernel.ErrForbidden, "The admin password is reset on the server only (see docs/accounts.md).")

// ErrTooManyInvites is returned at the limit of open invitations.
var ErrTooManyInvites = kernel.NewPublicError(kernel.ErrConflict, "At most 20 invitations can be open at the same time.")

// ErrNameTaken is returned when an invitation names an existing account or
// another open invitation.
var ErrNameTaken = kernel.NewPublicError(kernel.ErrConflict, "This name is already used by an account or an open invitation.")

// ErrNoPassword is returned for accounts without a password (kiosks).
var ErrNoPassword = kernel.NewPublicError(kernel.ErrInvalid, "This account has no password.")

// NormaliseEmail validates an address an account gives for reset links;
// "" means none. Only a bare address is accepted, no display name.
func NormaliseEmail(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}
	a, err := mail.ParseAddress(s)
	if err != nil || a.Name != "" || a.Address != s || len(s) > maxEmailLength || strings.ContainsAny(s, ",\r\n") {
		return "", kernel.NewValidationError("email", "Please enter a valid e-mail address.")
	}
	return s, nil
}

// ErrResetUnavailable is returned when reset mails cannot be sent.
var ErrResetUnavailable = kernel.NewPublicError(kernel.ErrUnavailable, "Password reset by e-mail is not set up on this instance. Ask the admin for a reset link.")

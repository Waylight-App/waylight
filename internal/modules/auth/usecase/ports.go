package usecase

import (
	"context"
	"time"

	"waylight/internal/modules/auth/domain"
)

// Input ports.

// GetStatus reports whether setup is required and whether the caller is logged in.
type GetStatus interface {
	Execute(ctx context.Context, req StatusRequest) (StatusResponse, error)
}

// SetupAccount creates the single account on first start and logs the user in.
type SetupAccount interface {
	Execute(ctx context.Context, req SetupRequest) (SessionResponse, error)
}

// Login verifies the password and starts a session.
type Login interface {
	Execute(ctx context.Context, req LoginRequest) (SessionResponse, error)
}

// Logout ends a session.
type Logout interface {
	Execute(ctx context.Context, req LogoutRequest) error
}

// Authenticate validates a session token for a protected request.
type Authenticate interface {
	Execute(ctx context.Context, req AuthenticateRequest) (AuthenticateResponse, error)
}

// ChangePassword replaces the password and ends all other sessions.
type ChangePassword interface {
	Execute(ctx context.Context, req ChangePasswordRequest) (SessionResponse, error)
}

// ManageSessions lists and revokes the sessions of the account.
type ManageSessions interface {
	List(ctx context.Context, currentToken string) ([]SessionView, error)
	Revoke(ctx context.Context, currentToken, sessionID string) error
	RevokeOthers(ctx context.Context, currentToken string) (int64, error)
}

// ResetAccount deletes the account and all sessions so setup runs again.
type ResetAccount interface {
	Execute(ctx context.Context) error
}

// PurgeExpiredSessions removes expired sessions from storage.
type PurgeExpiredSessions interface {
	Execute(ctx context.Context) (int64, error)
}

// Output ports.

// PreferenceRepository stores per-account preferences (ADR 0018).
type PreferenceRepository interface {
	PrivateByDefault(ctx context.Context, accountID string) (bool, error)
	SetPrivateByDefault(ctx context.Context, accountID string, on bool) error
}

// GroupRepository stores groups and who is in them (#294).
type GroupRepository interface {
	// List returns all groups ordered by name.
	List(ctx context.Context) ([]domain.Group, error)
	// Get returns domain.ErrGroupNotFound for an unknown group.
	Get(ctx context.Context, id string) (domain.Group, error)
	// Create and Update return domain.ErrGroupNameTaken for a duplicate name.
	Create(ctx context.Context, g domain.Group) error
	Update(ctx context.Context, g domain.Group) error
	Delete(ctx context.Context, id string) error
	// GroupsOf returns the IDs of the groups an account is in.
	GroupsOf(ctx context.Context, accountID string) ([]string, error)
	// RemoveAccount takes a deleted account out of every group.
	RemoveAccount(ctx context.Context, accountID string) error
}

// AccountRepository persists the owner, guest, kiosk and member accounts.
type AccountRepository interface {
	// Get returns domain.ErrAccountNotFound for a missing owner and
	// domain.ErrGuestNotFound for an unknown guest.
	Get(ctx context.Context, id string) (domain.Account, error)
	// FindByName returns domain.ErrGuestNotFound for unknown names.
	FindByName(ctx context.Context, name string) (domain.Account, error)
	// List returns all accounts, the owner first.
	List(ctx context.Context) ([]domain.Account, error)
	// Create returns domain.ErrAlreadySetUp for a second owner and
	// domain.ErrGuestNameTaken for a duplicate guest name.
	Create(ctx context.Context, a domain.Account) error
	Update(ctx context.Context, a domain.Account) error
	Delete(ctx context.Context, id string) error
}

// SessionRepository persists sessions.
type SessionRepository interface {
	Create(ctx context.Context, s domain.Session) error
	// Find returns domain.ErrSessionNotFound for unknown hashes.
	Find(ctx context.Context, hash domain.TokenHash) (domain.Session, error)
	Update(ctx context.Context, s domain.Session) error
	Delete(ctx context.Context, hash domain.TokenHash) error
	DeleteAll(ctx context.Context) error
	DeleteAllFor(ctx context.Context, accountID string) error
	DeleteExpired(ctx context.Context, now time.Time) (int64, error)
	// ListActive returns the account's sessions that have not expired at now.
	ListActive(ctx context.Context, accountID string, now time.Time) ([]domain.Session, error)
	// DeleteByPublicID returns domain.ErrSessionNotListed for unknown IDs.
	DeleteByPublicID(ctx context.Context, accountID, publicID string) error
	DeleteAllExcept(ctx context.Context, accountID string, keep domain.TokenHash) (int64, error)
}

// PairingCodeRepository stores the pairing codes of app sessions (#559).
type PairingCodeRepository interface {
	Create(ctx context.Context, c domain.PairingCode, now time.Time) error
	// Take removes the code and returns it; unknown, used and expired codes
	// yield domain.ErrInvalidPairingCode.
	Take(ctx context.Context, hash domain.TokenHash, now time.Time) (domain.PairingCode, error)
}

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) (bool, error)
}

// TokenGenerator creates unguessable random tokens.
type TokenGenerator interface {
	NewToken() (string, error)
}

// SetupTokenVerifier checks the one-time token that authorises the first setup.
type SetupTokenVerifier interface {
	Verify(token string) bool
}

// SetupObserver is notified once the account has been created, e.g. to fill
// the installation with sample data. It must not fail the setup.
type SetupObserver interface {
	// AccountCreated gets the start-page profile chosen in the setup (#281).
	AccountCreated(ctx context.Context, profile string)
}

// OneTimePasswords implements TOTP (RFC 6238) and recovery codes.
type OneTimePasswords interface {
	NewSecret() (string, error)
	ProvisioningURI(secret, account, issuer string) string
	// Verify returns the time step of a valid code.
	Verify(secret, code string, now time.Time) (int64, bool)
	NewRecoveryCode() (string, error)
}

// TwoFactor manages the second login factor (input port).
type TwoFactor interface {
	Status(ctx context.Context) (TwoFactorStatus, error)
	Start(ctx context.Context) (TwoFactorEnrollment, error)
	Confirm(ctx context.Context, code string) ([]string, error)
	Disable(ctx context.Context, req DisableTwoFactorRequest) error
}

// AttemptLimiter throttles repeated failed attempts per key (client IP).
type AttemptLimiter interface {
	Allow(key string, now time.Time) bool
	// RecordFailure reports whether this failure locked the client out.
	RecordFailure(key string, now time.Time) bool
	Reset(key string)
}

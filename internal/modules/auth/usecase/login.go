package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"waylight/internal/modules/auth/domain"
)

type login struct{ deps Deps }

// NewLogin returns the Login use case.
func NewLogin(deps Deps) Login { return &login{deps: deps} }

// Execute implements Login.
func (uc *login) Execute(ctx context.Context, req LoginRequest) (SessionResponse, error) {
	account, client, method, err := uc.verify(ctx, req)
	if err != nil {
		return SessionResponse{}, err
	}
	uc.deps.Events.loginSucceeded(ctx, account.ID(), client, method)
	return uc.deps.issuer().issue(ctx, account.ID(), client)
}

// verify checks the credentials and the second factor of req with rate
// limiting and the security log, and returns the account they belong to and
// the sign-in method. Browser and app sessions share it (#559).
func (uc *login) verify(ctx context.Context, req LoginRequest) (domain.Account, domain.Client, string, error) {
	fail := func(err error) (domain.Account, domain.Client, string, error) {
		return domain.Account{}, domain.Client{}, "", err
	}
	now := uc.deps.Clock.Now()
	if !uc.deps.Limiter.Allow(req.ClientKey, now) {
		return fail(domain.ErrTooManyAttempts)
	}
	owner, exists, err := loadAccount(ctx, uc.deps.Accounts)
	if err != nil {
		return fail(err)
	}
	if !exists {
		return fail(domain.ErrSetupRequired)
	}
	client := domain.NewClient(req.UserAgent, req.ClientKey)
	account, err := uc.account(ctx, owner, req.Username)
	if err != nil {
		return fail(err)
	}
	directory, checked, err := uc.viaDirectory(ctx, req)
	if err != nil {
		uc.deps.Events.loginFailed(ctx, account.ID(), client, uc.deps.Limiter.RecordFailure(req.ClientKey, now))
		return fail(domain.ErrInvalidCredentials)
	}
	if checked {
		account = directory
	}
	if account.IsOwner() && req.Username == "" && uc.deps.PasswordLoginDisabled {
		return fail(domain.ErrLocalLoginDisabled)
	}
	// Unknown and disabled guests are checked against the owner's hash so
	// the response time does not reveal which names exist.
	ok := checked
	if !checked {
		ok, err = uc.deps.Hasher.Verify(req.Password, account.PasswordHash())
		if err != nil {
			return fail(fmt.Errorf("verify password: %w", err))
		}
	}
	if !ok || account.Disabled() || (req.Username != "" && account.IsOwner()) {
		uc.deps.Events.loginFailed(ctx, account.ID(), client, uc.deps.Limiter.RecordFailure(req.ClientKey, now))
		return fail(domain.ErrInvalidCredentials)
	}
	keys, err := uc.deps.passkeysOf(ctx, account.ID())
	if err != nil {
		return fail(fmt.Errorf("list passkeys: %w", err))
	}
	if len(keys) > 0 && strings.TrimSpace(req.TOTPCode) == "" {
		// Second step: sign a challenge with a passkey (or send a code).
		return fail(uc.deps.startPasskeyLogin(account, keys))
	}
	if account.TOTPEnabled() || len(keys) > 0 {
		if err := uc.deps.checkSecondFactor(ctx, &account, req.TOTPCode, now); err != nil {
			if errors.Is(err, domain.ErrInvalidTOTPCode) {
				uc.deps.Events.loginFailed(ctx, account.ID(), client, uc.deps.Limiter.RecordFailure(req.ClientKey, now))
			}
			return fail(err)
		}
		if err := uc.deps.Accounts.Update(ctx, account); err != nil {
			return fail(fmt.Errorf("update account: %w", err))
		}
	}
	uc.deps.Limiter.Reset(req.ClientKey)
	method := "password"
	if checked {
		method = "directory"
	}
	return account, client, method, nil
}

// account resolves the login name: empty means the owner; an unknown name
// yields the owner account so the password check fails in constant time.
func (uc *login) account(ctx context.Context, owner domain.Account, name string) (domain.Account, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return owner, nil
	}
	a, err := uc.deps.Accounts.FindByName(ctx, name)
	// Kiosk displays have no password; they sign in with their link (#50).
	if errors.Is(err, domain.ErrGuestNotFound) || (err == nil && (a.IsOwner() || a.Role() == domain.RoleKiosk)) {
		return owner, nil
	}
	return a, err
}

// viaDirectory asks the directory about a named login. checked is true when
// the directory confirmed the password and named a member; false (with no
// error) means the name is checked locally as before. The owner (empty name)
// never goes to the directory.
func (uc *login) viaDirectory(ctx context.Context, req LoginRequest) (domain.Account, bool, error) {
	dir := uc.deps.Directory.get()
	name := strings.ToLower(strings.TrimSpace(req.Username))
	if dir == nil || name == "" {
		return domain.Account{}, false, nil
	}
	id, err := dir.Authenticate(ctx, name, req.Password)
	if errors.Is(err, ErrNotInDirectory) {
		return domain.Account{}, false, nil
	}
	if err != nil {
		return domain.Account{}, false, err
	}
	a, err := uc.deps.Accounts.Get(ctx, id)
	if err != nil {
		return domain.Account{}, false, err
	}
	// The directory may only sign in members and admins, never the owner or
	// a guest.
	if !a.IsPerson() {
		return domain.Account{}, false, domain.ErrInvalidCredentials
	}
	return a, true, nil
}

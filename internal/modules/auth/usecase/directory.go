package usecase

import (
	"context"
	"errors"
	"sync"
)

// ErrNotInDirectory tells the login to check the name locally instead: the
// directory does not know it. Any other error from Authenticate refuses the
// sign-in.
var ErrNotInDirectory = errors.New("the name is not in the directory")

// DirectoryAuthenticator checks a name and password against a directory such
// as LDAP or Active Directory (output port, #396). It returns the member
// account to sign in, creating or updating it as needed. It is set only in
// builds with the ee/ directory sync; the owner always signs in locally, so a
// directory that is down never locks the owner out.
type DirectoryAuthenticator interface {
	Authenticate(ctx context.Context, name, password string) (accountID string, err error)
}

// DirectoryLogin holds the authenticator once the server is built. Deps keep
// a pointer to it, so the login use case sees an authenticator set later.
type DirectoryLogin struct {
	mu   sync.Mutex
	impl DirectoryAuthenticator
}

// Set installs the authenticator.
func (d *DirectoryLogin) Set(a DirectoryAuthenticator) {
	d.mu.Lock()
	d.impl = a
	d.mu.Unlock()
}

func (d *DirectoryLogin) get() DirectoryAuthenticator {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.impl
}

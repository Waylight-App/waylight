// Package usecase signs members in with their directory password (#396).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"trisse-home/ee/ldap/domain"
	licence "trisse-home/ee/licence/domain"
	authdomain "trisse-home/internal/modules/auth/domain"
)

// Directory looks users up and checks passwords (output port).
type Directory interface {
	// Lookup finds the user with a login name; found = false when there is none.
	Lookup(ctx context.Context, name string) (user domain.User, found bool, err error)
	// Verify binds as the user; an error means a wrong password or a failure.
	Verify(ctx context.Context, dn, password string) error
}

// Link connects a login name with the member it signs in as.
type Link struct {
	Name          string
	AccountID     string
	Groups        []string // directory groups synced last time
	DisabledByDir bool     // locked out by this module, not by an admin
}

// Links stores links (output port).
type Links interface {
	Get(ctx context.Context, name string) (Link, bool, error)
	Save(ctx context.Context, l Link) error
	Delete(ctx context.Context, name string) error
}

// roleSource marks an admin role as granted by the directory, so it is taken back
// only by the directory (auth's domain.RoleSourceDirectory).
const roleSource = "directory"

// Accounts manages members in the auth module (output port).
type Accounts interface {
	CreateMember(ctx context.Context, name string) (string, error)
	UpdateGroups(ctx context.Context, accountID string, join, leave []string) error
	SetDisabled(ctx context.Context, accountID string, disabled bool) error
	// SetAdmin grants or takes back the admin role the provider gave (#434).
	SetAdmin(ctx context.Context, accountID string, admin bool, source string) (bool, error)
}

// Licensing tells how far directory sync may be used (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

// ErrNotInDirectory is returned when the login should be checked locally;
// module.go maps it to the auth module's sentinel of the same name.
var ErrNotInDirectory = errors.New("not in the directory")

// ErrRefused refuses a sign-in; the auth module answers it like any wrong
// password.
var ErrRefused = errors.New("the directory refused the sign-in")

// Deps are the ports.
type Deps struct {
	Directory Directory
	Links     Links
	Accounts  Accounts
	Licensing Licensing
	Policy    domain.Policy
}

// Authenticator implements the auth module's DirectoryAuthenticator.
type Authenticator struct{ d Deps }

// New returns the authenticator.
func New(d Deps) *Authenticator { return &Authenticator{d: d} }

// Authenticate checks a login against the directory. Without a licence key
// every name is checked locally, so linked members (random local password)
// cannot sign in; a name the directory does not know, or a directory that is
// down while the name is not linked, is checked locally too.
func (a *Authenticator) Authenticate(ctx context.Context, name, password string) (string, error) {
	level := a.d.Licensing.Level(ctx, licence.FeatureDirectorySync)
	if level == licence.LevelNone {
		return "", ErrNotInDirectory
	}
	link, linked, err := a.d.Links.Get(ctx, name)
	if err != nil {
		return "", fmt.Errorf("load link: %w", err)
	}
	user, found, err := a.d.Directory.Lookup(ctx, name)
	if err != nil {
		if linked {
			return "", fmt.Errorf("directory: %w", err)
		}
		return "", ErrNotInDirectory
	}
	if !found {
		if linked {
			// Removed from the directory: the member loses access, their data stays.
			if err := a.lockOut(ctx, link, level); err != nil {
				return "", err
			}
			return "", ErrRefused
		}
		return "", ErrNotInDirectory
	}
	if err := a.d.Directory.Verify(ctx, user.DN, password); err != nil {
		return "", ErrRefused
	}
	if !a.d.Policy.Allows(user) {
		if linked {
			if err := a.lockOut(ctx, link, level); err != nil {
				return "", err
			}
		}
		return "", ErrRefused
	}
	if linked {
		accountID, err := a.existing(ctx, link, user, level)
		if !errors.Is(err, authdomain.ErrGuestNotFound) {
			return accountID, err
		}
		if err := a.d.Links.Delete(ctx, name); err != nil { // an admin deleted the member
			return "", fmt.Errorf("delete link: %w", err)
		}
	}
	if level != licence.LevelFull {
		return "", ErrRefused
	}
	accountID, err := a.d.Accounts.CreateMember(ctx, name)
	if err != nil {
		return "", fmt.Errorf("create member: %w", err)
	}
	link = Link{Name: name, AccountID: accountID}
	return accountID, a.sync(ctx, &link, user)
}

func (a *Authenticator) lockOut(ctx context.Context, link Link, level licence.Level) error {
	if level != licence.LevelFull || link.DisabledByDir {
		return nil
	}
	if err := a.d.Accounts.SetDisabled(ctx, link.AccountID, true); err != nil && !errors.Is(err, authdomain.ErrGuestNotFound) {
		return err
	}
	link.DisabledByDir = true
	return a.d.Links.Save(ctx, link)
}

func (a *Authenticator) existing(ctx context.Context, link Link, user domain.User, level licence.Level) (string, error) {
	if level != licence.LevelFull {
		return link.AccountID, nil
	}
	if link.DisabledByDir {
		if err := a.d.Accounts.SetDisabled(ctx, link.AccountID, false); err != nil {
			return "", err
		}
		link.DisabledByDir = false
	}
	if err := a.sync(ctx, &link, user); err != nil {
		return "", err
	}
	return link.AccountID, nil
}

func (a *Authenticator) sync(ctx context.Context, link *Link, user domain.User) error {
	now := a.d.Policy.GroupsToSync(user)
	var leave []string
	for _, g := range link.Groups {
		if !containsFold(now, g) {
			leave = append(leave, g)
		}
	}
	if err := a.d.Accounts.UpdateGroups(ctx, link.AccountID, now, leave); err != nil {
		return err
	}
	if _, err := a.d.Accounts.SetAdmin(ctx, link.AccountID, a.d.Policy.IsAdmin(user), roleSource); err != nil {
		return err
	}
	link.Groups = now
	if err := a.d.Links.Save(ctx, *link); err != nil {
		return fmt.Errorf("save link: %w", err)
	}
	return nil
}

func containsFold(list []string, s string) bool {
	return slices.ContainsFunc(list, func(x string) bool { return strings.EqualFold(x, s) })
}

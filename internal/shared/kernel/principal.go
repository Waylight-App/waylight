package kernel

import (
	"context"
	"slices"
)

// Principal is the authenticated account of a request. The auth module puts
// it into the request context; other modules read it, e.g. to show a guest
// only the tabs shared with it.
type Principal struct {
	AccountID string
	// Role is "owner", "admin" (#434), "member", "guest", "kiosk" (#50), "token", or "system" for
	// work without a request (background jobs).
	Role string
	// Tabs are the dashboard tabs a guest may see (empty for the owner).
	Tabs []string
	// Scopes of an API token such as "todos:write" (Role "token").
	Scopes []string
	// Groups the account belongs to (#294); records shared with one of
	// them are visible to it (ADR 0018).
	Groups []string
	// PrivateByDefault is the admin's choice that their own new records
	// start private (ADR 0018, migration step 3). Members always start
	// private.
	PrivateByDefault bool
	// SecondFactorMissing is set when an admin policy requires a second
	// factor and the account has none (#304): until it sets one up, it may
	// only reach its own account settings.
	SecondFactorMissing bool
}

// SystemAccountID is the account that background jobs act for. It is the
// owner's ID, so records they create belong to the owner as before.
const SystemAccountID = "owner"

// AdminPrincipal is the first admin (the owner) as a principal. Code that
// works on instance-level data on behalf of the admin, such as config as
// code, uses it instead of the system, which would see every account's
// records (ADR 0018).
func AdminPrincipal() Principal { return Principal{AccountID: SystemAccountID, Role: "owner"} }

// IsSystem reports whether there is no signed-in principal: a background
// job, a CLI command or a test. The system sees all records (ADR 0018).
func (p Principal) IsSystem() bool { return p.Role == "system" }

// IsToken reports whether the request is authenticated by an API token.
func (p Principal) IsToken() bool { return p.Role == "token" }

// IsGuest reports whether the principal is a read-only guest.
func (p Principal) IsGuest() bool { return p.Role == "guest" }

// IsKiosk reports whether the principal is a read-only kiosk display (#50).
func (p Principal) IsKiosk() bool { return p.Role == "kiosk" }

// TabScoped reports whether the principal only sees some tabs (guests and
// kiosk displays).
func (p Principal) TabScoped() bool { return p.IsGuest() || p.IsKiosk() }

// CanSeeTab reports whether the principal may see a dashboard tab.
func (p Principal) CanSeeTab(id string) bool { return !p.TabScoped() || slices.Contains(p.Tabs, id) }

type principalKey struct{}

// WithPrincipal returns ctx carrying p.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFrom returns the principal of ctx. Work without one (jobs, CLI
// commands, tests) acts as the system, which writes as the owner and sees
// every record.
func PrincipalFrom(ctx context.Context) Principal {
	if p, ok := ctx.Value(principalKey{}).(Principal); ok {
		return p
	}
	return Principal{AccountID: SystemAccountID, Role: "system"}
}

// SettingsAccount is the account whose personal settings (ADR 0018) apply
// to p. Guests, kiosk displays and work without a request see the admin's,
// as the household look; everyone else has their own. API tokens act for
// their account.
func (p Principal) SettingsAccount() string {
	if p.TabScoped() || p.IsSystem() || p.AccountID == "" {
		return SystemAccountID
	}
	return p.AccountID
}

// IsAdmin reports whether p manages the instance: the owner, an admin
// (ADR 0022) or work without a request.
func (p Principal) IsAdmin() bool { return p.Role == "owner" || p.Role == "admin" || p.IsSystem() }

// PrincipalFor is the principal a background job uses when it acts for one
// account, for example to apply that account's settings to its records.
func PrincipalFor(accountID string) Principal {
	if accountID == SystemAccountID {
		return AdminPrincipal()
	}
	return Principal{AccountID: accountID, Role: "member"}
}

// ErrAdminOnly refuses a change that affects the whole instance to anyone
// but an admin, in places where the module otherwise belongs to everyone.
var ErrAdminOnly = NewPublicError(ErrForbidden, "Only an admin can change this.")

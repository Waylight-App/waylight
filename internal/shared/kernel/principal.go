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
	// Role is "owner", "guest", "kiosk" (#50), "token", or "system" for
	// work without a request (background jobs).
	Role string
	// Tabs are the dashboard tabs a guest may see (empty for the owner).
	Tabs []string
	// Scopes of an API token such as "todos:write" (Role "token").
	Scopes []string
	// Groups the account belongs to (#294); records shared with one of
	// them are visible to it (ADR 0018).
	Groups []string
}

// SystemAccountID is the account that background jobs act for. It is the
// owner's ID, so records they create belong to the owner as before.
const SystemAccountID = "owner"

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

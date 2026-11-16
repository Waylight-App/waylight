// Package domain holds the connected accounts of members (#326): one
// connection per member and service, made with OAuth on their behalf.
//
// Ring: Entities.
package domain

import "time"

// Status of a connection.
type Status string

// The statuses. Expired means the service no longer accepts the refresh
// token (revoked or timed out); the member has to connect again.
const (
	StatusConnected Status = "connected"
	StatusExpired   Status = "expired"
)

// Token is what the service issued; it never leaves the server.
type Token struct {
	Access  string
	Refresh string
	Expiry  time.Time // zero = unknown
}

// Connection is a member's account at one service.
type Connection struct {
	AccountID   string
	Service     string
	Token       Token
	Status      Status
	ConnectedAt time.Time
}

// RefreshMargin renews access tokens shortly before they expire.
const RefreshMargin = 2 * time.Minute

// NeedsRefresh reports whether the access token should be renewed at now.
func (c Connection) NeedsRefresh(now time.Time) bool {
	return !c.Token.Expiry.IsZero() && !now.Add(RefreshMargin).Before(c.Token.Expiry)
}

// Client is an admin's app registration at a service.
type Client struct {
	Service  string
	ClientID string
	Secret   string
	// Tenant is the Microsoft Entra tenant (ID or domain); other services ignore it.
	Tenant string
}

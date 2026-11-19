package kernel

import "context"

// AppSessions tells other modules whether an app session still exists
// (ADR 0024). The auth module implements it; the notifications module uses
// it to stop pushing to devices whose session was revoked or expired.
type AppSessions interface {
	AppSessionActive(ctx context.Context, accountID, publicID string) (bool, error)
}

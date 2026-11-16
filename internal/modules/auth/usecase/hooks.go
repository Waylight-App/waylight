package usecase

import (
	"context"
	"sync"
)

// AccountHooks tells other modules that an account was deleted, so they can
// remove what belongs only to it, such as the ee/ connected accounts' tokens
// (#326). Hooks run after the account is gone; an error there is theirs to
// log, the deletion stands.
type AccountHooks struct {
	mu      sync.Mutex
	deleted []func(ctx context.Context, accountID string)
}

// OnDeleted adds a hook.
func (h *AccountHooks) OnDeleted(fn func(ctx context.Context, accountID string)) {
	h.mu.Lock()
	h.deleted = append(h.deleted, fn)
	h.mu.Unlock()
}

func (h *AccountHooks) accountDeleted(ctx context.Context, accountID string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	fns := append([]func(context.Context, string){}, h.deleted...)
	h.mu.Unlock()
	for _, fn := range fns {
		fn(ctx, accountID)
	}
}

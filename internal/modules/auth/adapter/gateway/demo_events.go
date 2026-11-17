package gateway

import (
	"context"
	"time"

	"waylight/internal/modules/auth/domain"
)

// DiscardedEvents is the security log of the public demo. Every visitor
// signs in to the same account, so a real log would show each visitor the
// addresses and browsers of the others; this one keeps nothing. Every
// device counts as known, so no "new sign-in" alert is raised either.
type DiscardedEvents struct{}

// Add implements usecase.SecurityEventRepository and drops the event.
func (DiscardedEvents) Add(context.Context, domain.SecurityEvent) error { return nil }

// List implements usecase.SecurityEventRepository; the log is always empty.
func (DiscardedEvents) List(context.Context, int) ([]domain.SecurityEvent, error) { return nil, nil }

// KnownDevice implements usecase.SecurityEventRepository.
func (DiscardedEvents) KnownDevice(context.Context, string, string) (bool, error) { return true, nil }

// DeleteBefore implements usecase.SecurityEventRepository.
func (DiscardedEvents) DeleteBefore(context.Context, time.Time) (int64, error) { return 0, nil }

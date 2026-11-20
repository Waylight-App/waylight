package stream

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/pushrelay"
)

// Defaults of the mailbox (ADR 0027).
const (
	// MaxPerDevice is how many undelivered messages a device keeps.
	MaxPerDevice = 100
	// TTL is how long an undelivered message is kept.
	TTL = 72 * time.Hour
)

// Sender is the pushrelay.Sender of the platform "sse": it stores the
// message in the device's mailbox and wakes its stream if one is open.
type Sender struct {
	Mailbox Mailbox
	Hub     *Hub
	Now     func() time.Time
}

// ValidToken implements pushrelay.TokenChecker.
func (s *Sender) ValidToken(token string) bool { return ValidDeviceID(token) }

// Send implements pushrelay.Sender.
func (s *Sender) Send(ctx context.Context, token string, payload []byte) error {
	if !ValidDeviceID(token) {
		return pushrelay.ErrInvalid
	}
	if _, err := s.Mailbox.Put(ctx, token, payload, s.Now(), MaxPerDevice); err != nil {
		return fmt.Errorf("store message: %w", err)
	}
	s.Hub.Notify(token)
	return nil
}

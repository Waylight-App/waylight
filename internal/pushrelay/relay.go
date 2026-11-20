package pushrelay

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"waylight/internal/shared/pushproto"
)

// Platform names a push service.
type Platform string

// The ways the relay delivers: its own Server-Sent Events stream (ADR 0027)
// and Apple's push service for the iOS app.
const (
	SSE  Platform = "sse"
	APNs Platform = "apns"
)

// Message is one request of an instance: an opaque, sealed payload for one
// device token.
type Message struct {
	Platform Platform
	Token    string
	Payload  []byte
}

var (
	// ErrInvalid means the request is incomplete or names no known platform.
	ErrInvalid = errors.New("invalid message")
	// ErrTooLarge means the payload exceeds pushproto.MaxPayload.
	ErrTooLarge = errors.New("payload too large")
	// ErrForeignToken means the token is bound to another instance.
	ErrForeignToken = errors.New("token bound to another instance")
	// ErrRateLimited means a rate limit was reached.
	ErrRateLimited = errors.New("rate limited")
	// ErrUnregistered means the push service no longer accepts the token.
	ErrUnregistered = errors.New("token unregistered")
	// ErrUpstream means the push service failed for another reason.
	ErrUpstream = errors.New("push service failed")
)

// Sender delivers a payload through one push service. It returns
// ErrUnregistered for tokens the service rejects for good.
type Sender interface {
	Send(ctx context.Context, token string, payload []byte) error
}

// TokenChecker is implemented by senders that can tell a malformed device
// token without delivering, so no binding is created for it.
type TokenChecker interface {
	ValidToken(token string) bool
}

// TokenHash is the stored form of a device token.
type TokenHash [sha256.Size]byte

// Bindings remembers which instance key sent to a device token first.
type Bindings interface {
	// Bind binds hash to key unless it is bound already and returns the
	// bound key; it records the use either way.
	Bind(ctx context.Context, hash TokenHash, key string, now time.Time) (bound string, err error)
	Delete(ctx context.Context, hash TokenHash) error
}

// Limiter counts events per key and reports whether one more is allowed.
type Limiter interface {
	Allow(key string, now time.Time) bool
}

// Limits are the three rate limits of ADR 0025.
type Limits struct {
	Instance, Token, Client Limiter
}

// Relay is the use case.
type Relay struct {
	senders  map[Platform]Sender
	bindings Bindings
	limits   Limits
	now      func() time.Time
}

// New returns the relay. Platforms without a sender are rejected as
// invalid, so a relay can run without APNs.
func New(senders map[Platform]Sender, bindings Bindings, limits Limits, now func() time.Time) *Relay {
	return &Relay{senders: senders, bindings: bindings, limits: limits, now: now}
}

// Send checks limits and the token binding and forwards the message. key is
// the verified identity of the instance, client its network address.
func (r *Relay) Send(ctx context.Context, key, client string, m Message) error {
	sender, ok := r.senders[m.Platform]
	if !ok || m.Token == "" || len(m.Payload) == 0 {
		return ErrInvalid
	}
	if c, isChecker := sender.(TokenChecker); isChecker && !c.ValidToken(m.Token) {
		return ErrInvalid
	}
	if len(m.Payload) > pushproto.MaxPayload {
		return ErrTooLarge
	}
	now := r.now()
	hash := TokenHash(sha256.Sum256([]byte(string(m.Platform) + ":" + m.Token)))
	// The client limit comes first: it is the only one that cannot be
	// avoided by making up new keys or tokens.
	if !r.limits.Client.Allow(client, now) || !r.limits.Instance.Allow(key, now) || !r.limits.Token.Allow(string(hash[:]), now) {
		return ErrRateLimited
	}
	bound, err := r.bindings.Bind(ctx, hash, key, now)
	if err != nil {
		return fmt.Errorf("binding: %w", err)
	}
	if bound != key {
		return ErrForeignToken
	}
	switch err := sender.Send(ctx, m.Token, m.Payload); {
	case errors.Is(err, ErrUnregistered):
		if derr := r.bindings.Delete(ctx, hash); derr != nil {
			return errors.Join(err, derr)
		}
		return err
	case err != nil:
		return fmt.Errorf("%w: %w", ErrUpstream, err)
	}
	return nil
}

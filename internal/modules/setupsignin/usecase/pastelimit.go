package usecase

import (
	"time"

	"waylight/internal/shared/kernel"
)

// The page of a phone without the app is public and the code is its only
// credential, so wrong codes are limited per client.
const (
	missLimit  = 10
	missWindow = 5 * time.Minute
	// maxMissClients bounds the memory; at the bound every client waits.
	maxMissClients = 10000
)

var errTooManyGuesses = kernel.NewPublicError(kernel.ErrTooManyRequests, "Too many wrong codes. Please wait a few minutes.")

type miss struct {
	count int
	until time.Time
}

// missLimiter counts wrong codes per client in a fixed window. Callers hold
// the lock of the pastes.
type missLimiter struct{ clients map[string]*miss }

func newMissLimiter() *missLimiter { return &missLimiter{clients: map[string]*miss{}} }

func (l *missLimiter) allow(client string, now time.Time) bool {
	for k, m := range l.clients {
		if !now.Before(m.until) {
			delete(l.clients, k)
		}
	}
	if m := l.clients[client]; m != nil {
		return m.count < missLimit
	}
	return len(l.clients) < maxMissClients
}

func (l *missLimiter) fail(client string, now time.Time) {
	m := l.clients[client]
	if m == nil {
		m = &miss{until: now.Add(missWindow)}
		l.clients[client] = m
	}
	m.count++
}

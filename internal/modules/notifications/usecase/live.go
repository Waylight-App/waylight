package usecase

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/shared/kernel"
)

// Bounds of the live stream (#583).
const (
	// liveBacklog is how many recent events a reconnecting stream can
	// catch up on with Last-Event-ID.
	liveBacklog = 50
	// liveBuffer is the queue per stream; a stream that falls behind is
	// closed and catches up after reconnecting.
	liveBuffer = 16
	// MaxLiveStreams bounds the open streams of the instance.
	MaxLiveStreams = 32
)

// LiveEvent is one notification on the live stream.
type LiveEvent struct {
	// ID is "<boot>-<sequence>"; IDs of an earlier process run match no
	// event, so a stream after a restart starts without a backlog.
	ID      string
	At      time.Time
	Message domain.Message
}

// Live delivers the messages of the browser channel to the open web
// interface (#583). It holds no state beyond memory: notifications are
// the instance's (ADR 0018) and the channels keep their own history.
type Live interface {
	// Publisher receives the messages of domain.ChannelBrowser.
	Publisher
	// Subscribe opens a stream for an admin. backlog holds the events
	// after lastID; events is closed when the stream falls behind or
	// cancel is called.
	Subscribe(ctx context.Context, lastID string) (backlog []LiveEvent, events <-chan LiveEvent, cancel func(), err error)
}

// errTooManyStreams answers a stream beyond MaxLiveStreams.
var errTooManyStreams = kernel.NewPublicError(kernel.ErrUnavailable, "Too many open notification streams; close a tab and try again.")

type liveHub struct {
	clock kernel.Clock
	boot  string

	mu   sync.Mutex
	seq  uint64
	ring []LiveEvent
	subs map[chan LiveEvent]struct{}
}

// NewLive returns the in-memory hub of the live stream.
func NewLive(clock kernel.Clock) Live {
	return &liveHub{
		clock: clock,
		boot:  strconv.FormatInt(clock.Now().UnixNano(), 36),
		subs:  map[chan LiveEvent]struct{}{},
	}
}

// Publish implements Publisher: the event goes to every open stream.
func (h *liveHub) Publish(_ context.Context, _ domain.Channel, m domain.Message) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++
	e := LiveEvent{ID: h.boot + "-" + strconv.FormatUint(h.seq, 10), At: h.clock.Now(), Message: m}
	h.ring = append(h.ring, e)
	if len(h.ring) > liveBacklog {
		h.ring = h.ring[len(h.ring)-liveBacklog:]
	}
	for ch := range h.subs {
		select {
		case ch <- e:
		default:
			// Too slow: close it; the browser reconnects and catches up.
			delete(h.subs, ch)
			close(ch)
		}
	}
	return nil
}

func (h *liveHub) Subscribe(ctx context.Context, lastID string) ([]LiveEvent, <-chan LiveEvent, func(), error) {
	if p := kernel.PrincipalFrom(ctx); !p.IsAdmin() || p.IsSystem() {
		return nil, nil, nil, kernel.ErrAdminOnly
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.subs) >= MaxLiveStreams {
		return nil, nil, nil, errTooManyStreams
	}
	ch := make(chan LiveEvent, liveBuffer)
	h.subs[ch] = struct{}{}
	cancel := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if _, ok := h.subs[ch]; ok {
			delete(h.subs, ch)
			close(ch)
		}
	}
	return h.after(lastID), ch, cancel, nil
}

// after returns the events after id; an unknown or empty id gives none,
// so a new tab does not replay old notifications.
func (h *liveHub) after(id string) []LiveEvent {
	if id == "" || !strings.HasPrefix(id, h.boot+"-") {
		return nil
	}
	for i, e := range h.ring {
		if e.ID == id {
			return append([]LiveEvent(nil), h.ring[i+1:]...)
		}
	}
	// Older than the backlog: everything still kept.
	n, err := strconv.ParseUint(strings.TrimPrefix(id, h.boot+"-"), 10, 64)
	if err != nil || n >= h.seq {
		return nil
	}
	return append([]LiveEvent(nil), h.ring...)
}

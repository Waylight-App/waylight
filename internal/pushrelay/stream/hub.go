package stream

import "sync"

// Subscription is the open stream of one device.
type Subscription struct {
	// Wake receives a signal when new messages are waiting.
	Wake <-chan struct{}
	// Replaced is closed when the same device opened a newer stream.
	Replaced <-chan struct{}

	wake     chan struct{}
	replaced chan struct{}
}

// Hub tracks the open streams and wakes them when messages arrive
// (observer). Each device has at most one stream; a new one replaces the
// old, so a stale connection never keeps a device from receiving.
type Hub struct {
	mu   sync.Mutex
	subs map[string]*Subscription
}

// NewHub returns an empty hub.
func NewHub() *Hub { return &Hub{subs: map[string]*Subscription{}} }

// Subscribe registers the stream of device, replacing an older one, and
// returns it with a function that removes it again.
func (h *Hub) Subscribe(device string) (*Subscription, func()) {
	s := &Subscription{wake: make(chan struct{}, 1), replaced: make(chan struct{})}
	s.Wake, s.Replaced = s.wake, s.replaced
	h.mu.Lock()
	if old, ok := h.subs[device]; ok {
		close(old.replaced)
	}
	h.subs[device] = s
	h.mu.Unlock()
	return s, func() {
		h.mu.Lock()
		if h.subs[device] == s {
			delete(h.subs, device)
		}
		h.mu.Unlock()
	}
}

// Notify wakes the stream of device, if one is open. It never blocks: a
// pending signal already means "read the mailbox".
func (h *Hub) Notify(device string) {
	h.mu.Lock()
	s := h.subs[device]
	h.mu.Unlock()
	if s == nil {
		return
	}
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Open reports how many streams are open.
func (h *Hub) Open() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs)
}

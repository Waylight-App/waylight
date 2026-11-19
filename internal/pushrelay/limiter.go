package pushrelay

import (
	"sync"
	"time"
)

// Window is a fixed-window limiter: at most Max events per key in each
// window. Windows are kept in memory only (ADR 0025).
type Window struct {
	Max    int
	Period time.Duration

	mu     sync.Mutex
	start  time.Time
	counts map[string]int
}

// Allow counts one event for key and reports whether it is within the limit.
func (w *Window) Allow(key string, now time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.counts == nil || now.Sub(w.start) >= w.Period {
		// A new window drops every count, which also bounds the memory used.
		w.start, w.counts = now, map[string]int{}
	}
	if w.counts[key] >= w.Max {
		return false
	}
	w.counts[key]++
	return true
}

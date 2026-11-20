package stream

import "sync"

// Gate bounds the open streams per client address and in total, so a
// single address cannot hold all connections of the relay.
type Gate struct {
	MaxPerClient int
	MaxTotal     int

	mu     sync.Mutex
	total  int
	counts map[string]int
}

// Acquire reserves a stream for client and reports whether one was free.
func (g *Gate) Acquire(client string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.counts == nil {
		g.counts = map[string]int{}
	}
	if g.total >= g.MaxTotal || g.counts[client] >= g.MaxPerClient {
		return false
	}
	g.total++
	g.counts[client]++
	return true
}

// Release frees a stream reserved by Acquire.
func (g *Gate) Release(client string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.total--
	if g.counts[client]--; g.counts[client] <= 0 {
		delete(g.counts, client)
	}
}

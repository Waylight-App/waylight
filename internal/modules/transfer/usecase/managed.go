package usecase

import (
	"slices"
	"sync"
)

// Managed is the set of sections that CONFIG_FILE owns (#88). It is filled
// once at start-up and read by the config and import use cases, which
// leave those sections alone.
type Managed struct {
	mu  sync.RWMutex
	set map[string]bool
}

// Set replaces the managed sections.
func (m *Managed) Set(sections []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.set = map[string]bool{}
	for _, s := range sections {
		m.set[s] = true
	}
}

// Has reports whether a section is managed.
func (m *Managed) Has(section string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.set[section]
}

// List returns the managed sections, sorted.
func (m *Managed) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.set))
	for s := range m.set {
		out = append(out, s)
	}
	slices.Sort(out)
	return out
}

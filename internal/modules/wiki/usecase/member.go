package usecase

import (
	"context"
	"errors"
	"sync"

	"waylight/internal/modules/wiki/domain"
)

// ErrNotConnected tells the search to skip a member source quietly: the
// searching person has not connected that service.
var ErrNotConnected = errors.New("the searching person has not connected this service")

// MemberSource searches a wiki on behalf of the signed-in person (output
// port, #327–#329), such as Confluence with that person's own account, so
// results never show what they may not see. Only the ee/ team integrations
// add sources; the core searches the wikis the admin connected.
type MemberSource interface {
	// Name is shown with the results, e.g. "Confluence".
	Name() string
	// Search reads the person from the context. It returns ErrNotConnected
	// when the person has no account at the service.
	Search(ctx context.Context, query string, limit int) ([]domain.Page, error)
}

// MemberSources holds the sources once the server is built; Deps keep a
// pointer so sources added later are searched too.
type MemberSources struct {
	mu   sync.Mutex
	list []MemberSource
}

// Add adds a source.
func (m *MemberSources) Add(s MemberSource) {
	m.mu.Lock()
	m.list = append(m.list, s)
	m.mu.Unlock()
}

func (m *MemberSources) all() []MemberSource {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MemberSource{}, m.list...)
}

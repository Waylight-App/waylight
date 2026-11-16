// Package usecase answers the Teams command box search for the member
// behind the request (#334, ADR 0023).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"strings"
	"sync"
	"time"

	licence "trisse-home/ee/licence/domain"
	"trisse-home/ee/teamsbot/domain"
)

// Connections finds the members who connected Microsoft 365 (output port).
type Connections interface {
	AccountsWith(ctx context.Context, service string) ([]string, error)
	AccessToken(ctx context.Context, accountID, service string) (string, error)
}

// Directory tells whose a Microsoft token is: the Entra object ID (output
// port, Graph /me).
type Directory interface {
	ObjectID(ctx context.Context, token string) (string, error)
}

// Finder searches the start page as an account (output port).
type Finder interface {
	Find(ctx context.Context, accountID, query string) ([]domain.Hit, error)
}

// Licensing tells whether chat apps are licensed (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

// Deps are the collaborators.
type Deps struct {
	Connections Connections
	Directory   Directory
	Finder      Finder
	Licensing   Licensing
	Now         func() time.Time
}

const (
	rescanAfter = 30 * time.Second
	scanWorkers = 8
	service     = "microsoft"
)

// Service implements the use case.
type Service struct {
	d        Deps
	mu       sync.Mutex
	members  map[string]string // Entra object ID → account
	scanned  time.Time
	scanning sync.Mutex
}

// New returns the use case.
func New(d Deps) *Service { return &Service{d: d, members: map[string]string{}} }

// Query answers a search typed into the command box by the Teams user with
// the Entra object ID objectID.
func (s *Service) Query(ctx context.Context, objectID, query string) domain.Response {
	if s.d.Licensing.Level(ctx, licence.FeatureChatApps) != licence.LevelFull {
		return domain.Message(domain.Unlicensed)
	}
	account, ok := s.member(ctx, strings.ToLower(objectID))
	if !ok {
		return domain.Message(domain.NotConnected)
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return domain.Message("Type what you are looking for: a service, a bookmark or a wiki page.")
	}
	hits, err := s.d.Finder.Find(ctx, account, q)
	if err != nil && len(hits) == 0 {
		return domain.Message(domain.Failed)
	}
	return domain.Results(q, hits)
}

func (s *Service) member(ctx context.Context, objectID string) (string, bool) {
	if objectID == "" {
		return "", false
	}
	s.mu.Lock()
	acc, ok := s.members[objectID]
	stale := s.d.Now().Sub(s.scanned) >= rescanAfter
	s.mu.Unlock()
	if ok || !stale {
		return acc, ok
	}
	s.scan(ctx)
	s.mu.Lock()
	defer s.mu.Unlock()
	acc, ok = s.members[objectID]
	return acc, ok
}

// Warm fills the map before the first query.
func (s *Service) Warm(ctx context.Context) { s.scan(ctx) }

// scan asks Graph for the object ID behind every connected Microsoft
// account, in parallel, and replaces the map.
func (s *Service) scan(ctx context.Context) {
	s.scanning.Lock()
	defer s.scanning.Unlock()
	accounts, err := s.d.Connections.AccountsWith(ctx, service)
	if err != nil {
		return
	}
	var mu sync.Mutex
	found := map[string]string{}
	jobs := make(chan string)
	var wg sync.WaitGroup
	for range min(scanWorkers, len(accounts)) {
		wg.Go(func() {
			for acc := range jobs {
				token, err := s.d.Connections.AccessToken(ctx, acc, service)
				if err != nil {
					continue
				}
				id, err := s.d.Directory.ObjectID(ctx, token)
				if err != nil || id == "" {
					continue
				}
				mu.Lock()
				found[strings.ToLower(id)] = acc
				mu.Unlock()
			}
		})
	}
	for _, acc := range accounts {
		jobs <- acc
	}
	close(jobs)
	wg.Wait()
	s.mu.Lock()
	s.members, s.scanned = found, s.d.Now()
	s.mu.Unlock()
}

// Package usecase searches team wikis with the signed-in person's own
// account (#327–#329).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"errors"
	"sync"
	"time"

	licence "waylight/ee/licence/domain"
	"waylight/internal/modules/wiki"
	"waylight/internal/modules/wiki/domain"
	"waylight/internal/shared/kernel"
)

// Tokens hands out a member's access token at a service (output port, the
// connected accounts of #326).
type Tokens interface {
	AccessToken(ctx context.Context, accountID, service string) (string, error)
}

// Licensing tells whether workplace search is licensed (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

// Searcher asks one service with a token (output port).
type Searcher interface {
	Search(ctx context.Context, token, query string, limit int) ([]domain.Page, error)
}

// Source is one service, searched for whoever asks. It implements
// wiki.MemberSource.
type Source struct {
	name, service string
	tokens        Tokens
	licensing     Licensing
	searcher      Searcher

	// perMinute, when set, bounds the searches per person and minute, below
	// the service's own limit (Slack allows about ten), so typing never gets
	// the person's account throttled at the service.
	perMinute int
	now       func() time.Time
	mu        sync.Mutex
	recent    map[string][]time.Time
}

// ErrBusy is returned when a person searched a source too often in the last
// minute; the search lists the source as failed, a visible hint.
var ErrBusy = errors.New("too many searches in the last minute")

// WithLimit bounds the searches per person and minute.
func (s *Source) WithLimit(perMinute int, now func() time.Time) *Source {
	s.perMinute, s.now, s.recent = perMinute, now, map[string][]time.Time{}
	return s
}

// allow records a search by who and reports whether it is within the limit.
func (s *Source) allow(who string) bool {
	if s.perMinute == 0 {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	kept := s.recent[who][:0]
	for _, t := range s.recent[who] {
		if now.Sub(t) < time.Minute {
			kept = append(kept, t)
		}
	}
	if len(kept) >= s.perMinute {
		s.recent[who] = kept
		return false
	}
	s.recent[who] = append(kept, now)
	return true
}

// NewSource returns a source named name (shown with the results) for a
// connected-accounts service.
func NewSource(name, service string, tokens Tokens, licensing Licensing, searcher Searcher) *Source {
	return &Source{name: name, service: service, tokens: tokens, licensing: licensing, searcher: searcher}
}

// Name implements wiki.MemberSource.
func (s *Source) Name() string { return s.name }

// Search implements wiki.MemberSource. Without a licence key, without a
// signed-in person or without their connection the source is skipped: the
// person sees the other results as before. An expired connection is
// skipped too; the connected accounts page asks them to reconnect.
func (s *Source) Search(ctx context.Context, query string, limit int) ([]domain.Page, error) {
	if s.licensing.Level(ctx, licence.FeatureWorkplaceSearch) == licence.LevelNone {
		return nil, wiki.ErrNotConnected
	}
	p := kernel.PrincipalFrom(ctx)
	// The system (a background job) is nobody: it must not search with the
	// owner's tokens, whose ID it carries.
	if p.IsSystem() || p.IsToken() || p.AccountID == "" {
		return nil, wiki.ErrNotConnected
	}
	who := p.AccountID
	token, err := s.tokens.AccessToken(ctx, who, s.service)
	if err != nil {
		var pe *kernel.PublicError
		if errors.As(err, &pe) {
			// Not connected, expired, or no licence for connections.
			return nil, wiki.ErrNotConnected
		}
		return nil, err
	}
	if !s.allow(who) {
		return nil, ErrBusy
	}
	return s.searcher.Search(ctx, token, query, limit)
}

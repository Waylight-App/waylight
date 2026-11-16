// Package usecase searches team wikis with the signed-in person's own
// account (#327–#329).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"errors"

	licence "trisse-home/ee/licence/domain"
	"trisse-home/internal/modules/wiki"
	"trisse-home/internal/modules/wiki/domain"
	"trisse-home/internal/shared/kernel"
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
	return s.searcher.Search(ctx, token, query, limit)
}

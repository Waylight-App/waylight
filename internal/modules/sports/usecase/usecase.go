// Package usecase holds the sports use cases (#285).
//
// Ring: Use Cases.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"trisse-home/internal/modules/sports/domain"
	"trisse-home/internal/shared/kernel"
)

// cacheTTL: results change on match days only; a quarter of an hour keeps
// the provider's load low (cache-aside, ADR 0006).
const cacheTTL = 15 * time.Minute

// TopRows is how many table rows the league view shows.
const TopRows = 5

// Provider reads a league (output port); OpenLigaDB is the first adapter.
type Provider interface {
	Teams(ctx context.Context, league string, season int) ([]domain.Team, error)
	Matches(ctx context.Context, league string, season int) ([]domain.Match, error)
	Table(ctx context.Context, league string, season int) ([]domain.Standing, error)
}

// Overview is what the widget shows.
type Overview struct {
	League   domain.League
	Season   int
	Team     *domain.Team
	Last     *domain.Match
	Next     *domain.Match
	Position *domain.Standing
	// Table is the top of the table (league view) or empty.
	Table []domain.Standing
}

// Sports is the input port.
type Sports interface {
	Leagues() []domain.League
	Teams(ctx context.Context, league string) ([]domain.Team, error)
	// Overview of a league, or of one of its teams when teamID > 0.
	Overview(ctx context.Context, league string, teamID int) (Overview, error)
}

type cached struct {
	at      time.Time
	teams   []domain.Team
	matches []domain.Match
	table   []domain.Standing
}

type sports struct {
	p     Provider
	clock kernel.Clock
	mu    sync.Mutex
	cache map[string]*cached
}

// New returns the use cases.
func New(p Provider, clock kernel.Clock) Sports {
	return &sports{p: p, clock: clock, cache: map[string]*cached{}}
}

func (uc *sports) Leagues() []domain.League { return domain.Leagues }

func unavailable(err error) error {
	var pe *kernel.PublicError
	if errors.As(err, &pe) {
		return err
	}
	return kernel.TargetFailure(kernel.ErrUnavailable, err)
}

// load returns the league data, refreshed after cacheTTL; a failing
// provider keeps the last known data.
func (uc *sports) load(ctx context.Context, league string) (*cached, int, error) {
	now := uc.clock.Now()
	season := domain.Season(now)
	key := fmt.Sprintf("%s/%d", league, season)
	uc.mu.Lock()
	c := uc.cache[key]
	uc.mu.Unlock()
	if c != nil && now.Sub(c.at) < cacheTTL {
		return c, season, nil
	}
	teams, err := uc.p.Teams(ctx, league, season)
	var matches []domain.Match
	var table []domain.Standing
	if err == nil {
		matches, err = uc.p.Matches(ctx, league, season)
	}
	if err == nil {
		table, err = uc.p.Table(ctx, league, season)
	}
	if err != nil {
		if c != nil {
			return c, season, nil
		}
		return nil, season, unavailable(err)
	}
	c = &cached{at: now, teams: teams, matches: matches, table: table}
	uc.mu.Lock()
	uc.cache[key] = c
	uc.mu.Unlock()
	return c, season, nil
}

func (uc *sports) Teams(ctx context.Context, league string) ([]domain.Team, error) {
	l, err := domain.ParseLeague(league)
	if err != nil {
		return nil, err
	}
	// The provider only ever gets an ID from the fixed list, never the input.
	c, _, err := uc.load(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	return c.teams, nil
}

func (uc *sports) Overview(ctx context.Context, league string, teamID int) (Overview, error) {
	l, err := domain.ParseLeague(league)
	if err != nil {
		return Overview{}, err
	}
	c, season, err := uc.load(ctx, l.ID)
	if err != nil {
		return Overview{}, err
	}
	o := Overview{League: l, Season: season}
	if teamID <= 0 {
		o.Table = c.table[:min(TopRows, len(c.table))]
		return o, nil
	}
	for i := range c.teams {
		if c.teams[i].ID == teamID {
			o.Team = &c.teams[i]
		}
	}
	if o.Team == nil {
		return Overview{}, kernel.NewPublicError(kernel.ErrNotFound, "The team does not play in this league this season.")
	}
	o.Last, o.Next = domain.LastAndNext(c.matches, teamID, uc.clock.Now())
	o.Position = domain.PositionOf(c.table, teamID)
	return o, nil
}

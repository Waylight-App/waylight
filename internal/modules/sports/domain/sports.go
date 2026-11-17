// Package domain holds the sports rules (#285).
//
// Ring: Entities.
package domain

import (
	"slices"
	"time"

	"waylight/internal/shared/kernel"
)

// League is a competition a provider serves.
type League struct {
	ID   string // provider shortcut, e.g. "bl1"
	Name string
}

// Leagues lists the competitions of the first provider (OpenLigaDB).
var Leagues = []League{
	{ID: "bl1", Name: "1. Bundesliga"},
	{ID: "bl2", Name: "2. Bundesliga"},
	{ID: "bl3", Name: "3. Liga"},
}

// ErrUnknownLeague rejects leagues no provider serves.
var ErrUnknownLeague = kernel.NewValidationError("league", "Choose one of the offered leagues.")

// ParseLeague validates a league ID.
func ParseLeague(id string) (League, error) {
	i := slices.IndexFunc(Leagues, func(l League) bool { return l.ID == id })
	if i < 0 {
		return League{}, ErrUnknownLeague
	}
	return Leagues[i], nil
}

// Season is the football season that is running at t: it starts in July,
// so November 2026 is season 2026 (2026/27) and March 2027 too.
func Season(t time.Time) int {
	if t.Month() >= time.July {
		return t.Year()
	}
	return t.Year() - 1
}

// Team is a club of a league.
type Team struct {
	ID        int
	Name      string
	ShortName string
}

// Match is one fixture or result.
type Match struct {
	Kickoff    time.Time
	Home, Away Team
	Finished   bool
	// HomeGoals and AwayGoals are the final score (or the current one).
	HomeGoals, AwayGoals int
	HasScore             bool
	Matchday             string
}

// Standing is one row of a table.
type Standing struct {
	Position int
	Team     Team
	Played   int
	Points   int
	GoalDiff int
}

// LastAndNext returns the most recent finished match and the next open one
// of a team (zero = none), by kickoff time.
func LastAndNext(matches []Match, teamID int, now time.Time) (last, next *Match) {
	for i := range matches {
		m := &matches[i]
		if m.Home.ID != teamID && m.Away.ID != teamID {
			continue
		}
		switch {
		case m.Finished:
			if last == nil || m.Kickoff.After(last.Kickoff) {
				last = m
			}
		case !m.Kickoff.Before(now.Add(-3 * time.Hour)): // a running match counts as next
			if next == nil || m.Kickoff.Before(next.Kickoff) {
				next = m
			}
		}
	}
	return last, next
}

// PositionOf returns the table row of a team (nil = not in the table).
func PositionOf(table []Standing, teamID int) *Standing {
	for i := range table {
		if table[i].Team.ID == teamID {
			return &table[i]
		}
	}
	return nil
}

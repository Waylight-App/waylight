// Package presenter maps the sports overview to JSON (#285).
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/sports/domain"
	"waylight/internal/modules/sports/usecase"
)

// TeamDTO is a club.
type TeamDTO struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
}

// MatchDTO is a fixture or result; goals are set when hasScore.
type MatchDTO struct {
	Kickoff   string  `json:"kickoff"`
	Home      TeamDTO `json:"home"`
	Away      TeamDTO `json:"away"`
	Finished  bool    `json:"finished"`
	HasScore  bool    `json:"hasScore"`
	HomeGoals int     `json:"homeGoals"`
	AwayGoals int     `json:"awayGoals"`
	Matchday  string  `json:"matchday"`
}

// StandingDTO is a table row.
type StandingDTO struct {
	Position int     `json:"position"`
	Team     TeamDTO `json:"team"`
	Played   int     `json:"played"`
	Points   int     `json:"points"`
	GoalDiff int     `json:"goalDiff"`
}

// LeagueDTO is a competition.
type LeagueDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// OverviewDTO is returned by GET /api/sports/overview.
type OverviewDTO struct {
	League   LeagueDTO     `json:"league"`
	Season   int           `json:"season"`
	Team     *TeamDTO      `json:"team"`
	Last     *MatchDTO     `json:"last"`
	Next     *MatchDTO     `json:"next"`
	Position *StandingDTO  `json:"position"`
	Table    []StandingDTO `json:"table"`
}

// Team maps a team.
func Team(t domain.Team) TeamDTO { return TeamDTO{ID: t.ID, Name: t.Name, ShortName: t.ShortName} }

// Teams maps teams.
func Teams(ts []domain.Team) []TeamDTO {
	out := make([]TeamDTO, 0, len(ts))
	for _, t := range ts {
		out = append(out, Team(t))
	}
	return out
}

// Leagues maps leagues.
func Leagues(ls []domain.League) []LeagueDTO {
	out := make([]LeagueDTO, 0, len(ls))
	for _, l := range ls {
		out = append(out, LeagueDTO(l))
	}
	return out
}

func match(m *domain.Match) *MatchDTO {
	if m == nil {
		return nil
	}
	return &MatchDTO{Kickoff: m.Kickoff.Format(time.RFC3339), Home: Team(m.Home), Away: Team(m.Away), Finished: m.Finished,
		HasScore: m.HasScore, HomeGoals: m.HomeGoals, AwayGoals: m.AwayGoals, Matchday: m.Matchday}
}

func standing(s domain.Standing) StandingDTO {
	return StandingDTO{Position: s.Position, Team: Team(s.Team), Played: s.Played, Points: s.Points, GoalDiff: s.GoalDiff}
}

// Overview maps the use case answer.
func Overview(o usecase.Overview) OverviewDTO {
	out := OverviewDTO{League: LeagueDTO(o.League), Season: o.Season, Last: match(o.Last), Next: match(o.Next), Table: make([]StandingDTO, 0, len(o.Table))}
	if o.Team != nil {
		t := Team(*o.Team)
		out.Team = &t
	}
	if o.Position != nil {
		p := standing(*o.Position)
		out.Position = &p
	}
	for _, s := range o.Table {
		out.Table = append(out.Table, standing(s))
	}
	return out
}

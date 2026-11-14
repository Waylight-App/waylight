// Package gateway reads football leagues from OpenLigaDB (#285).
//
// Ring: Interface Adapters.
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"trisse-home/internal/modules/sports/domain"
)

// DefaultOpenLigaDBURL is the public OpenLigaDB API (no API key).
const DefaultOpenLigaDBURL = "https://api.openligadb.de"

// OpenLigaDB implements usecase.Provider.
type OpenLigaDB struct {
	client *http.Client
	base   string
}

// NewOpenLigaDB returns the provider.
func NewOpenLigaDB(client *http.Client, base string) *OpenLigaDB {
	if base == "" {
		base = DefaultOpenLigaDBURL
	}
	return &OpenLigaDB{client: client, base: strings.TrimSuffix(base, "/")}
}

func (o *OpenLigaDB) get(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.base+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenLigaDB answered %s", res.Status)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(dst); err != nil {
		return fmt.Errorf("parse OpenLigaDB answer: %w", err)
	}
	return nil
}

type olTeam struct {
	TeamID    int    `json:"teamId"`
	TeamName  string `json:"teamName"`
	ShortName string `json:"shortName"`
}

func (t olTeam) team() domain.Team {
	short := t.ShortName
	if short == "" {
		short = t.TeamName
	}
	return domain.Team{ID: t.TeamID, Name: t.TeamName, ShortName: short}
}

func path(parts ...string) string {
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return "/" + strings.Join(parts, "/")
}

// Teams implements usecase.Provider.
func (o *OpenLigaDB) Teams(ctx context.Context, league string, season int) ([]domain.Team, error) {
	var raw []olTeam
	if err := o.get(ctx, path("getavailableteams", league, strconv.Itoa(season)), &raw); err != nil {
		return nil, err
	}
	out := make([]domain.Team, 0, len(raw))
	for _, t := range raw {
		out = append(out, t.team())
	}
	return out, nil
}

// Matches implements usecase.Provider: every match of the season.
func (o *OpenLigaDB) Matches(ctx context.Context, league string, season int) ([]domain.Match, error) {
	var raw []struct {
		MatchDateTimeUTC string `json:"matchDateTimeUTC"`
		Team1            olTeam `json:"team1"`
		Team2            olTeam `json:"team2"`
		MatchIsFinished  bool   `json:"matchIsFinished"`
		MatchResults     []struct {
			ResultTypeID int `json:"resultTypeID"`
			PointsTeam1  int `json:"pointsTeam1"`
			PointsTeam2  int `json:"pointsTeam2"`
		} `json:"matchResults"`
		Group struct {
			GroupName string `json:"groupName"`
		} `json:"group"`
	}
	if err := o.get(ctx, path("getmatchdata", league, strconv.Itoa(season)), &raw); err != nil {
		return nil, err
	}
	out := make([]domain.Match, 0, len(raw))
	for _, m := range raw {
		kickoff, err := time.Parse(time.RFC3339, m.MatchDateTimeUTC)
		if err != nil {
			continue
		}
		dm := domain.Match{Kickoff: kickoff, Home: m.Team1.team(), Away: m.Team2.team(), Finished: m.MatchIsFinished, Matchday: m.Group.GroupName}
		// resultTypeID 2 is the final result, 1 the half-time score.
		best := -1
		for _, r := range m.MatchResults {
			if r.ResultTypeID > best {
				best, dm.HomeGoals, dm.AwayGoals, dm.HasScore = r.ResultTypeID, r.PointsTeam1, r.PointsTeam2, true
			}
		}
		out = append(out, dm)
	}
	return out, nil
}

// Table implements usecase.Provider.
func (o *OpenLigaDB) Table(ctx context.Context, league string, season int) ([]domain.Standing, error) {
	var raw []struct {
		TeamInfoID int    `json:"teamInfoId"`
		TeamName   string `json:"teamName"`
		ShortName  string `json:"shortName"`
		Points     int    `json:"points"`
		Matches    int    `json:"matches"`
		GoalDiff   int    `json:"goalDiff"`
	}
	if err := o.get(ctx, path("getbltable", league, strconv.Itoa(season)), &raw); err != nil {
		return nil, err
	}
	out := make([]domain.Standing, 0, len(raw))
	for i, r := range raw {
		out = append(out, domain.Standing{Position: i + 1, Team: olTeam{TeamID: r.TeamInfoID, TeamName: r.TeamName, ShortName: r.ShortName}.team(),
			Played: r.Matches, Points: r.Points, GoalDiff: r.GoalDiff})
	}
	return out, nil
}

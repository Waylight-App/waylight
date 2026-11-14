package domain

import (
	"testing"
	"time"
)

func TestSeasonAndLeague(t *testing.T) {
	if Season(time.Date(2026, 11, 14, 0, 0, 0, 0, time.UTC)) != 2026 || Season(time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)) != 2026 {
		t.Fatal("season")
	}
	if l, err := ParseLeague("bl2"); err != nil || l.Name != "2. Bundesliga" {
		t.Fatal(l, err)
	}
	if _, err := ParseLeague("pl"); err == nil {
		t.Fatal("unknown league accepted")
	}
}

func TestLastAndNext(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	bayern, bvb, svw := Team{ID: 40}, Team{ID: 7}, Team{ID: 134}
	ms := []Match{
		{Kickoff: now.AddDate(0, 0, -14), Home: bayern, Away: svw, Finished: true},
		{Kickoff: now.AddDate(0, 0, -7), Home: bvb, Away: bayern, Finished: true, HomeGoals: 1, AwayGoals: 2, HasScore: true},
		{Kickoff: now.AddDate(0, 0, 14), Home: bayern, Away: bvb},
		{Kickoff: now.AddDate(0, 0, 7), Home: svw, Away: bayern},
		{Kickoff: now.AddDate(0, 0, 1), Home: bvb, Away: svw},
	}
	last, next := LastAndNext(ms, 40, now)
	if last == nil || last.Home.ID != 7 || next == nil || next.Home.ID != 134 {
		t.Fatalf("%+v %+v", last, next)
	}
	if l, n := LastAndNext(ms, 999, now); l != nil || n != nil {
		t.Fatal("unknown team")
	}
	table := []Standing{{Position: 1, Team: bayern}, {Position: 2, Team: bvb}}
	if PositionOf(table, 7).Position != 2 || PositionOf(table, 134) != nil {
		t.Fatal("position")
	}
}

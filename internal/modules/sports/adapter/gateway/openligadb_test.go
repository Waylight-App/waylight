package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Answers as OpenLigaDB sends them (shortened).
var recorded = map[string]string{
	"/getavailableteams/bl1/2026": `[{"teamId":40,"teamName":"FC Bayern München","shortName":"Bayern","teamIconUrl":"https://x","teamGroupName":null},{"teamId":7,"teamName":"Borussia Dortmund","shortName":"BVB"}]`,
	"/getmatchdata/bl1/2026": `[{"matchID":1,"matchDateTime":"2026-11-08T15:30:00","timeZoneID":"W. Europe Standard Time","leagueShortcut":"bl1","matchDateTimeUTC":"2026-11-08T14:30:00Z",` +
		`"group":{"groupName":"10. Spieltag","groupOrderID":10},"team1":{"teamId":7,"teamName":"Borussia Dortmund","shortName":"BVB"},"team2":{"teamId":40,"teamName":"FC Bayern München","shortName":"Bayern"},` +
		`"matchIsFinished":true,"matchResults":[{"resultID":1,"resultName":"Halbzeit","pointsTeam1":0,"pointsTeam2":1,"resultTypeID":1},{"resultID":2,"resultName":"Endergebnis","pointsTeam1":1,"pointsTeam2":2,"resultTypeID":2}]},` +
		`{"matchID":2,"matchDateTimeUTC":"2026-11-21T14:30:00Z","group":{"groupName":"11. Spieltag"},"team1":{"teamId":40,"teamName":"FC Bayern München","shortName":""},"team2":{"teamId":7,"teamName":"Borussia Dortmund","shortName":"BVB"},"matchIsFinished":false,"matchResults":[]}]`,
	"/getbltable/bl1/2026": `[{"teamInfoId":40,"teamName":"FC Bayern München","shortName":"Bayern","points":28,"opponentGoals":8,"goals":30,"matches":10,"won":9,"lost":0,"draw":1,"goalDiff":22},{"teamInfoId":7,"teamName":"Borussia Dortmund","shortName":"BVB","points":21,"matches":10,"goalDiff":9}]`,
}

func TestOpenLigaDB(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := recorded[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()
	o := NewOpenLigaDB(srv.Client(), srv.URL)
	ctx := context.Background()
	teams, err := o.Teams(ctx, "bl1", 2026)
	if err != nil || len(teams) != 2 || teams[0].ShortName != "Bayern" {
		t.Fatalf("%+v %v", teams, err)
	}
	ms, err := o.Matches(ctx, "bl1", 2026)
	if err != nil || len(ms) != 2 || !ms[0].Finished || ms[0].HomeGoals != 1 || ms[0].AwayGoals != 2 || ms[0].Matchday != "10. Spieltag" {
		t.Fatalf("final score, not half-time: %+v %v", ms, err)
	}
	if ms[1].HasScore || ms[1].Home.ShortName != "FC Bayern München" {
		t.Fatalf("open match: %+v", ms[1])
	}
	table, err := o.Table(ctx, "bl1", 2026)
	if err != nil || len(table) != 2 || table[1].Position != 2 || table[0].Points != 28 || table[0].GoalDiff != 22 {
		t.Fatalf("%+v %v", table, err)
	}
	if _, err := o.Teams(ctx, "bl9", 2026); err == nil {
		t.Fatal("404 accepted")
	}
}

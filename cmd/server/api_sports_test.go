package main

import (
	"fmt"
	"net/http"
	"time"

	"waylight/internal/platform/config"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "sports",
		Upstream: func(mux *http.ServeMux) {
			// OpenLigaDB (#285), for the season that is running now.
			season := time.Now().Year()
			if time.Now().Month() < time.July {
				season--
			}
			mux.HandleFunc(fmt.Sprintf("GET /oldb/getavailableteams/bl1/%d", season), func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, `[{"teamId":40,"teamName":"FC Bayern München","shortName":"Bayern"},{"teamId":7,"teamName":"Borussia Dortmund","shortName":"BVB"}]`)
			})
			mux.HandleFunc(fmt.Sprintf("GET /oldb/getmatchdata/bl1/%d", season), func(w http.ResponseWriter, _ *http.Request) {
				past := time.Now().AddDate(0, 0, -3).UTC().Format(time.RFC3339)
				next := time.Now().AddDate(0, 0, 4).UTC().Format(time.RFC3339)
				fmt.Fprintf(w, `[{"matchDateTimeUTC":%q,"team1":{"teamId":7,"teamName":"Borussia Dortmund","shortName":"BVB"},"team2":{"teamId":40,"teamName":"FC Bayern München","shortName":"Bayern"},"matchIsFinished":true,"matchResults":[{"resultTypeID":2,"pointsTeam1":1,"pointsTeam2":3}]},`+
					`{"matchDateTimeUTC":%q,"team1":{"teamId":40,"teamName":"FC Bayern München","shortName":"Bayern"},"team2":{"teamId":7,"teamName":"Borussia Dortmund","shortName":"BVB"},"matchIsFinished":false,"matchResults":[]}]`, past, next)
			})
			mux.HandleFunc(fmt.Sprintf("GET /oldb/getbltable/bl1/%d", season), func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprint(w, `[{"teamInfoId":40,"teamName":"FC Bayern München","shortName":"Bayern","points":28,"matches":10,"goalDiff":22},{"teamInfoId":7,"teamName":"Borussia Dortmund","shortName":"BVB","points":21,"matches":10,"goalDiff":9}]`)
			})
		},
		Config: func(cfg *config.Config, upstream string) { cfg.SportsAPIURL = upstream + "/oldb" },
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var teams []obj
			c.json(http.MethodGet, "/api/sports/teams?league=bl1", nil, http.StatusOK, &teams)
			if len(teams) != 2 {
				t.Fatalf("teams: %v", teams)
			}
			var o obj
			c.json(http.MethodGet, "/api/sports/overview?league=bl1&team=7", nil, http.StatusOK, &o)
			last, next, pos := o["last"].(obj), o["next"].(obj), o["position"].(obj)
			if last["awayGoals"] != float64(3) || next["home"].(obj)["shortName"] != "Bayern" || pos["position"] != float64(2) {
				t.Fatalf("overview: %v", o)
			}
			c.json(http.MethodGet, "/api/sports/overview?league=bl1", nil, http.StatusOK, &o)
			if len(o["table"].([]any)) != 2 {
				t.Fatalf("league: %v", o)
			}
			c.do(http.MethodGet, "/api/sports/overview?league=nba", nil, http.StatusUnprocessableEntity)
		},
	})
}

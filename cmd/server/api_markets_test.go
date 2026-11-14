package main

import (
	"fmt"
	"net/http"
	"time"

	"trisse-home/internal/platform/config"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "markets",
		Upstream: func(mux *http.ServeMux) {
			// ECB reference rates and Stooq closes (#284), dated relative to today.
			d1, d2 := time.Now().AddDate(0, 0, -2).Format(time.DateOnly), time.Now().AddDate(0, 0, -1).Format(time.DateOnly)
			mux.HandleFunc("GET /ecb.xml", func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(w, `<Envelope><Cube><Cube time=%q><Cube currency="USD" rate="1.10"/></Cube><Cube time=%q><Cube currency="USD" rate="1.00"/></Cube></Cube></Envelope>`, d2, d1)
			})
			mux.HandleFunc("GET /stooq/", func(w http.ResponseWriter, _ *http.Request) {
				fmt.Fprintf(w, "Date,Open,High,Low,Close,Volume\n%s,1,1,1,200,0\n%s,1,1,1,190,0\n", d1, d2)
			})
		},
		Config: func(cfg *config.Config, upstream string) {
			cfg.MarketsECBURL = upstream + "/ecb.xml"
			cfg.MarketsStooqURL = upstream + "/stooq/"
		},
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var rows []obj
			c.json(http.MethodGet, "/api/markets/watchlist?symbols=fx:USD,stooq:aapl.us", nil, http.StatusOK, &rows)
			if len(rows) != 2 || rows[0]["value"] != 1.1 || rows[0]["changePct"].(float64) < 9.99 || rows[1]["value"] != float64(190) {
				t.Fatalf("watchlist: %v", rows)
			}
			c.do(http.MethodGet, "/api/markets/watchlist?symbols=yahoo:aapl", nil, http.StatusUnprocessableEntity)
		},
	})
}

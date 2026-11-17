package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"waylight/internal/platform/config"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "electricity",
		Upstream: func(mux *http.ServeMux) {
			// aWATTar market data (#288): four quarter hours from the requested start.
			mux.HandleFunc("GET /awattar/de/v1/marketdata", func(w http.ResponseWriter, r *http.Request) {
				start, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
				fmt.Fprint(w, `{"object":"list","data":[`)
				for q := range 4 {
					s := start + int64(q)*int64(15*time.Minute/time.Millisecond)
					sep := ","
					if q == 3 {
						sep = ""
					}
					fmt.Fprintf(w, `{"start_timestamp":%d,"end_timestamp":%d,"marketprice":%d,"unit":"Eur/MWh"}%s`, s, s+900000, 100+q*10, sep)
				}
				fmt.Fprint(w, `]}`)
			})
		},
		Config: func(cfg *config.Config, upstream string) {
			cfg.ElectricityAPIURL = upstream + "/awattar/{tld}/v1"
		},
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var p obj
			c.json(http.MethodGet, "/api/electricity/prices?zone=DE-LU", nil, http.StatusOK, &p)
			hours, _ := p["hours"].([]any)
			if len(hours) != 1 || hours[0].(obj)["centPerKWh"] != 11.5 || hours[0].(obj)["cheapest"] != true || p["zone"] != "DE-LU" {
				t.Fatalf("electricity: %v", p)
			}
			c.do(http.MethodGet, "/api/electricity/prices?zone=FR", nil, http.StatusUnprocessableEntity)
		},
	})
}

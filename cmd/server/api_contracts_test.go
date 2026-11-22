package main

import (
	"net/http"
	"time"
)

// End-to-end check of the contracts module against the logged-in test
// server (#687).
func init() {
	registerModuleE2E(moduleE2E{
		Name: "contracts",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			end := time.Now().AddDate(0, 0, 40).Format(time.DateOnly)
			deadline := time.Now().AddDate(0, 0, 10).Format(time.DateOnly)
			var created obj
			c.json(http.MethodPost, "/api/contracts", obj{"name": "Gym", "costCents": 3600, "period": "quarterly", "termEnd": end,
				"noticeDays": 30, "renewMonths": 12, "remindDays": 14}, http.StatusCreated, &created)
			if created["deadline"] != deadline || created["soon"] != true || created["monthlyCents"] != float64(1200) {
				t.Fatalf("contracts: %v", created)
			}
			var compact obj
			c.json(http.MethodGet, "/api/contracts/compact", nil, http.StatusOK, &compact)
			if compact["level"] != "attention" {
				t.Fatalf("contracts compact: %v", compact)
			}
			id, _ := created["id"].(string)
			c.json(http.MethodPut, "/api/contracts/"+id, obj{"name": "Gym", "period": "weekly", "termEnd": end}, http.StatusUnprocessableEntity, nil)
			c.do(http.MethodDelete, "/api/contracts/"+id, nil, http.StatusNoContent)
		},
	})
}

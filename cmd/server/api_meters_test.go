package main

import (
	"net/http"
	"time"
)

// End-to-end check of the meters module against the logged-in test server
// (#689): create, readings, order check, consumption, compact view, remove.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "meters",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var created obj
			c.json(http.MethodPost, "/api/meters", obj{"name": "Power", "kind": "electricity", "unit": "kWh"}, http.StatusCreated, &created)
			id, _ := created["id"].(string)
			if created["visibility"] != "everyone" || created["last"] != nil {
				t.Fatalf("meters: %v", created)
			}
			today := time.Now()
			first, second := today.AddDate(0, 0, -20).Format(time.DateOnly), today.AddDate(0, 0, -10).Format(time.DateOnly)
			c.json(http.MethodPost, "/api/meters/"+id+"/readings", obj{"day": first, "value": 100}, http.StatusCreated, nil)
			c.json(http.MethodPost, "/api/meters/"+id+"/readings", obj{"day": second, "value": 90}, http.StatusUnprocessableEntity, nil)
			c.json(http.MethodPost, "/api/meters/"+id+"/readings", obj{"day": second, "value": 130}, http.StatusCreated, nil)
			var detail obj
			c.json(http.MethodGet, "/api/meters/"+id, nil, http.StatusOK, &detail)
			if rs, _ := detail["readings"].([]any); len(rs) != 2 || detail["consumption"] == nil {
				t.Fatalf("meters detail: %v", detail)
			}
			c.do(http.MethodGet, "/api/meters/"+id+"?year=abc", nil, http.StatusUnprocessableEntity)
			var compact obj
			c.json(http.MethodGet, "/api/meters/compact", nil, http.StatusOK, &compact)
			if compact["level"] != "none" {
				t.Fatalf("meters compact: %v", compact)
			}
			c.do(http.MethodDelete, "/api/meters/"+id+"/readings/"+second, nil, http.StatusNoContent)
			c.do(http.MethodDelete, "/api/meters/"+id, nil, http.StatusNoContent)
			c.do(http.MethodGet, "/api/meters/"+id, nil, http.StatusNotFound)
		},
	})
}

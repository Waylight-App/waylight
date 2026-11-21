package main

import "net/http"

// End-to-end check of service discovery (#590): an integration that
// declares how it is found answers with a list (nothing runs in the test
// network), an unknown one with 404.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "discovery",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var found []obj
			c.json(http.MethodGet, "/api/discovery/immich", nil, http.StatusOK, &found)
			if found == nil {
				t.Fatal("discovery answers with a list")
			}
			c.do(http.MethodGet, "/api/discovery/clock", nil, http.StatusNotFound)
		},
	})
}

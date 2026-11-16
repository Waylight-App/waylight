package main

import "net/http"

// End-to-end check of the phone list module (#300) against the logged-in
// test server: add, search with the dial number, update and delete.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "phonelist",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var created obj
			c.json(http.MethodPost, "/api/phonelist/entries", obj{"name": "Anna Weber", "role": "Reception", "phone": "+49 30 123"}, http.StatusCreated, &created)
			var hits []obj
			c.json(http.MethodGet, "/api/phonelist/entries?q=anna", nil, http.StatusOK, &hits)
			if len(hits) != 1 || hits[0]["dial"] != "+4930123" {
				t.Fatalf("phonelist search: %v", hits)
			}
			id := created["id"].(string)
			c.json(http.MethodPut, "/api/phonelist/entries/"+id, obj{"name": "Anna Weber", "email": "anna@example.org"}, http.StatusOK, nil)
			c.do(http.MethodDelete, "/api/phonelist/entries/"+id, nil, http.StatusNoContent)
		},
	})
}

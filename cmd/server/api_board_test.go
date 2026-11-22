package main

import "net/http"

// End-to-end check of the family board module against the logged-in test
// server (#685).
func init() {
	registerModuleE2E(moduleE2E{
		Name: "board",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var created obj
			c.json(http.MethodPost, "/api/board", obj{"text": "Dentist on Friday", "pinned": true}, http.StatusCreated, &created)
			if created["visibility"] != "everyone" || created["mine"] != true {
				t.Fatalf("board: %v", created)
			}
			var compact obj
			c.json(http.MethodGet, "/api/board/compact", nil, http.StatusOK, &compact)
			if title, _ := compact["title"].(map[string]any); title["text"] != "1 note" {
				t.Fatalf("board compact: %v", compact)
			}
			c.json(http.MethodPost, "/api/board", obj{"text": "x", "expiresOn": "2000-01-01"}, http.StatusUnprocessableEntity, nil)
			id, _ := created["id"].(string)
			c.do(http.MethodDelete, "/api/board/"+id, nil, http.StatusNoContent)
		},
	})
}

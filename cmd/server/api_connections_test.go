package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// End-to-end check of the connections module (#610): the credential is
// stored but never answered, a widget that names the connection keeps it
// from being removed without confirmation.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "connections",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var created obj
			c.json(http.MethodPost, "/api/connections", obj{"service": "immich", "name": "Home", "url": "http://192.168.178.20:2283",
				"credential": obj{"apiKey": "e2e-secret-key"}}, http.StatusCreated, &created)
			id := created["id"].(string)
			var tab obj
			c.json(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Connections"}, http.StatusCreated, &tab)
			c.json(http.MethodPost, "/api/dashboard/tabs/"+tab["id"].(string)+"/widgets",
				obj{"kind": "immich", "x": 0, "y": 0, "w": 4, "h": 2, "config": obj{"connectionId": id}}, http.StatusCreated, nil)
			raw := c.do(http.MethodGet, "/api/connections?service=immich", nil, http.StatusOK)
			var list []obj
			_ = json.Unmarshal(raw, &list)
			if len(list) != 1 || len(list[0]["usedBy"].([]any)) != 1 || strings.Contains(string(raw), "e2e-secret-key") {
				t.Fatalf("connections: %s", raw)
			}
			c.json(http.MethodPut, "/api/connections/"+id+"/credential", obj{"credential": obj{"apiKey": "other"}}, http.StatusOK, nil)
			c.do(http.MethodDelete, "/api/connections/"+id, nil, http.StatusConflict)
			c.do(http.MethodDelete, "/api/connections/"+id+"?confirm=true", nil, http.StatusNoContent)
			c.do(http.MethodDelete, "/api/dashboard/tabs/"+tab["id"].(string), nil, http.StatusNoContent)
		},
	})
}

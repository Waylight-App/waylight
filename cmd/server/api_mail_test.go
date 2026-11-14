package main

import "net/http"

func init() {
	registerModuleE2E(moduleE2E{
		Name: "mail",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			// No account yet: the widget says so; an unreachable server is refused before saving.
			c.do(http.MethodGet, "/api/mail/inbox", nil, http.StatusNotFound)
			var list []obj
			c.json(http.MethodGet, "/api/mail/settings/accounts", nil, http.StatusOK, &list)
			if len(list) != 0 {
				t.Fatalf("accounts: %v", list)
			}
			c.do(http.MethodPost, "/api/mail/settings/accounts", obj{"name": "x", "host": "bad host", "username": "u", "password": "p"}, http.StatusUnprocessableEntity)
			c.do(http.MethodPost, "/api/mail/settings/accounts", obj{"name": "x", "host": "127.0.0.1", "port": 1, "username": "u", "password": "p"}, http.StatusBadGateway)
		},
	})
}

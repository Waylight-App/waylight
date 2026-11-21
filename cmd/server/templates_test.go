package main

import (
	"net/http"
	"testing"
)

// TestBuiltinTemplates applies every built-in template on a server with
// the default modules (#589): each one exists, and no widget is left out,
// so a template never names a kind the server does not know.
func TestBuiltinTemplates(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var list []obj
	owner.json(http.MethodGet, "/api/dashboard/templates/builtin", nil, http.StatusOK, &list)
	if len(list) != 4 {
		t.Fatalf("templates: %+v", list)
	}
	for _, tpl := range list {
		var tab obj
		body := obj{"location": obj{"lat": 53.14, "lon": 8.21, "label": "Oldenburg"}}
		owner.json(http.MethodPost, "/api/dashboard/templates/builtin/"+tpl["id"].(string)+"/tabs", body, http.StatusCreated, &tab)
		if got, want := len(tab["widgets"].([]any)), len(tpl["widgets"].([]any)); got != want {
			t.Errorf("%s: %d of %d widgets", tpl["id"], got, want)
		}
	}
	owner.do(http.MethodPost, "/api/dashboard/templates/builtin/nope/tabs", obj{}, http.StatusNotFound)
}

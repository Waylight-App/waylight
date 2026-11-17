package main

import (
	"net/http"
	"testing"

	"waylight/internal/platform/config"
)

// TestSetupProfileEveryday checks the start-page profile of the first setup
// (#281): "everyday" seeds no Server tab and is stored; changing it later
// keeps every tab and widget.
func TestSetupProfileEveryday(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.SampleData = true })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123", "profile": "everyday"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var dash struct {
		Tabs []struct {
			Name string `json:"name"`
		} `json:"tabs"`
	}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 1 || dash.Tabs[0].Name != "Home" {
		t.Fatalf("everyday tabs: %+v", dash.Tabs)
	}
	var a obj
	c.json(http.MethodGet, "/api/appearance", nil, http.StatusOK, &a)
	if a["profile"] != "everyday" {
		t.Fatalf("profile: %v", a)
	}
	// Saving the appearance without a profile (older clients) keeps it.
	a["accentColor"] = "#123456"
	delete(a, "profile")
	c.json(http.MethodPut, "/api/appearance", a, http.StatusOK, &a)
	if a["profile"] != "everyday" {
		t.Fatalf("profile lost: %v", a)
	}
	// Switching to homelab changes nothing on the dashboard.
	a["profile"] = "homelab"
	c.json(http.MethodPut, "/api/appearance", a, http.StatusOK, &a)
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if a["profile"] != "homelab" || len(dash.Tabs) != 1 {
		t.Fatalf("after switching: %v %+v", a, dash.Tabs)
	}
	a["profile"] = "server-room"
	c.do(http.MethodPut, "/api/appearance", a, http.StatusUnprocessableEntity)
}

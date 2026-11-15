package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"trisse-home/internal/platform/config"
)

// TestCalendarAccounts checks the wiring of #282: without PUBLIC_URL signing
// in is unavailable, with it the start answers with the provider's sign-in
// address, and the callback refuses an unknown state.
func TestCalendarAccounts(t *testing.T) {
	up := fakeUpstream(t)
	public, _ := url.Parse("https://home.example.org")
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.PublicURL = public })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var list obj
	c.json(http.MethodGet, "/api/calendar/accounts", nil, http.StatusOK, &list)
	if list["redirectUrl"] != "https://home.example.org/api/calendar/accounts/oauth/callback" || len(list["accounts"].([]any)) != 0 {
		t.Fatalf("accounts: %v", list)
	}
	var start obj
	c.json(http.MethodPost, "/api/calendar/accounts/oauth/start", obj{"provider": "google", "clientId": "abc.apps.googleusercontent.com"}, http.StatusOK, &start)
	if u := start["url"].(string); !strings.HasPrefix(u, "https://accounts.google.com/") || !strings.Contains(u, "code_challenge=") {
		t.Fatalf("start: %v", start)
	}
	var status obj
	c.json(http.MethodGet, "/api/calendar/accounts/oauth/state/"+start["state"].(string), nil, http.StatusOK, &status)
	if status["status"] != "pending" {
		t.Fatalf("status: %v", status)
	}
	c.do(http.MethodPost, "/api/calendar/accounts/oauth/start", obj{"provider": "yahoo", "clientId": "x"}, http.StatusUnprocessableEntity)
	c.do(http.MethodGet, "/api/calendar/accounts/oauth/callback?state=forged&code=x", nil, http.StatusBadGateway)

	var writable []obj
	c.json(http.MethodGet, "/api/calendar/writable", nil, http.StatusOK, &writable)
	if len(writable) != 0 {
		t.Fatalf("writable: %v", writable)
	}
}

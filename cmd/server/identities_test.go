package main

import (
	"net/http"
	"strings"
	"testing"

	"waylight/internal/platform/config"
)

// TestLinkPersonalMicrosoftAccount links a sign-in at the provider to the
// owner while signed in (#527). Personal Microsoft accounts send no
// email_verified, so the e-mail list cannot let them in; the link can.
func TestLinkPersonalMicrosoftAccount(t *testing.T) {
	var appURL string
	idp := fakeIDP(t, "me@outlook.com", &appURL, map[string]any{"email_verified": nil})
	app := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) {
		c.OIDCIssuer, c.OIDCClientID, c.OIDCClientSecret = idp.URL, "waylight", "secret"
		c.OIDCAllowedEmails = []string{"me@outlook.com"}
		c.OIDCRedirectURL = "http://placeholder/api/auth/oidc/callback"
	})
	appURL = app.URL
	owner := newClient(t, app.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	// A fresh browser following the provider's redirects like a person would.
	signIn := func() bool {
		b := newClient(t, app.URL)
		b.http.CheckRedirect = nil
		res, err := b.http.Get(app.URL + "/api/auth/oidc/start")
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		var st obj
		b.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
		return st["authenticated"] == true
	}
	if signIn() {
		t.Fatal("signed in with an unverified e-mail address")
	}

	// Link while signed in: the page gets the provider URL and opens it.
	var start obj
	owner.json(http.MethodPost, "/api/auth/identities/link", nil, http.StatusOK, &start)
	owner.http.CheckRedirect = nil
	res, err := owner.http.Get(start["redirect"].(string))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if !strings.HasSuffix(res.Request.URL.String(), "/?settings=account&link=ok") {
		t.Fatalf("link ended at %s", res.Request.URL)
	}
	var links []obj
	owner.json(http.MethodGet, "/api/auth/identities", nil, http.StatusOK, &links)
	if len(links) != 1 || links[0]["email"] != "me@outlook.com" || links[0]["subject"] != "user-1" {
		t.Fatalf("links: %v", links)
	}

	if !signIn() {
		t.Fatal("the linked sign-in did not sign in")
	}

	// Linking the same identity again is refused, disconnecting ends it.
	owner.json(http.MethodPost, "/api/auth/identities/link", nil, http.StatusOK, &start)
	if res, err = owner.http.Get(start["redirect"].(string)); err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if !strings.HasSuffix(res.Request.URL.String(), "/?settings=account&link=taken") {
		t.Fatalf("second link ended at %s", res.Request.URL)
	}
	owner.do(http.MethodDelete, "/api/auth/identities/user-1", nil, http.StatusNoContent)
	if signIn() {
		t.Fatal("signed in after disconnecting")
	}
}

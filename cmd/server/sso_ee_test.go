//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"slices"
	"testing"
	"time"

	"waylight/ee/licence/domain"
	"waylight/internal/platform/config"
)

// TestSSOForMembers signs a member in through OpenID Connect in the ee build
// (#302): refused without a licence key, created with its group with one.
func TestSSOForMembers(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved, savedMembers := licencePublicKey, membersEnabled
	licencePublicKey, membersEnabled = pub, true
	t.Cleanup(func() { licencePublicKey, membersEnabled = saved, savedMembers })
	t.Setenv("OIDC_MEMBER_EMAILS", "@example.org")

	var appURL string
	idp := fakeIDP(t, "anna@example.org", &appURL, map[string]any{"preferred_username": "anna", "groups": []string{"Office"}})
	app := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) {
		c.OIDCIssuer, c.OIDCClientID, c.OIDCClientSecret = idp.URL, "waylight", "secret"
		c.OIDCAllowedEmails = []string{"owner@example.org"}
		c.OIDCRedirectURL = "http://placeholder/api/auth/oidc/callback"
		c.OIDCGroupsClaim = "groups"
	})
	appURL = app.URL
	owner := newClient(t, app.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	signIn := func() obj {
		b := newClient(t, app.URL)
		b.http.CheckRedirect = nil
		res, err := b.http.Get(app.URL + "/api/auth/oidc/start")
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		var st obj
		b.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
		return st
	}

	// Without a licence key only the owner signs in through the provider.
	if st := signIn(); st["authenticated"] == true {
		t.Fatalf("member signed in without a licence key: %v", st)
	}

	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"sso"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	if st := signIn(); st["authenticated"] != true || st["role"] != "member" || st["username"] != "anna" {
		t.Fatalf("member not signed in: %v", st)
	}
	var members struct {
		Members []obj `json:"members"`
	}
	owner.json(http.MethodGet, "/api/auth/members", nil, http.StatusOK, &members)
	if len(members.Members) != 1 || members.Members[0]["name"] != "anna" {
		t.Fatalf("members: %v", members.Members)
	}
	var groups []obj
	owner.json(http.MethodGet, "/api/auth/groups", nil, http.StatusOK, &groups)
	id := members.Members[0]["id"].(string)
	found := slices.ContainsFunc(groups, func(g obj) bool {
		accounts, _ := g["accounts"].([]any)
		return g["name"] == "Office" && slices.Contains(accounts, any(id))
	})
	if !found {
		t.Fatalf("groups: %v", groups)
	}
}

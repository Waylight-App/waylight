//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"testing"
	"time"

	"waylight/ee/licence/domain"
	"waylight/internal/platform/config"
)

// TestSignUpFromInvitationWithSSO accepts an invitation through the provider
// instead of a password (#528): offered only with a licence key that
// includes sso; the account is created, signed in and linked, and the link
// is its only way in until a password is set.
func TestSignUpFromInvitationWithSSO(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved, savedMembers := licencePublicKey, membersEnabled
	licencePublicKey, membersEnabled = pub, true
	t.Cleanup(func() { licencePublicKey, membersEnabled = saved, savedMembers })

	var appURL string
	idp := fakeIDP(t, "sam@outlook.com", &appURL, map[string]any{"email_verified": nil, "sub": "sam-sub"})
	app := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) {
		c.OIDCIssuer, c.OIDCClientID, c.OIDCClientSecret = idp.URL, "waylight", "secret"
		c.OIDCRedirectURL = "http://placeholder/api/auth/oidc/callback"
	})
	appURL = app.URL
	owner := newClient(t, app.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var created obj
	owner.json(http.MethodPost, "/api/auth/invites", obj{"name": "sam", "role": "member", "tabs": []string{}, "validDays": 7}, http.StatusCreated, &created)
	token := created["token"].(string)

	visitor := newClient(t, app.URL)
	var info obj
	visitor.json(http.MethodPost, "/api/auth/link/inspect", obj{"token": token}, http.StatusOK, &info)
	if info["sso"] == true {
		t.Fatalf("offered without a licence key: %v", info)
	}
	visitor.do(http.MethodPost, "/api/auth/link/sso", obj{"token": token}, http.StatusNotFound)

	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"sso"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	visitor.json(http.MethodPost, "/api/auth/link/inspect", obj{"token": token}, http.StatusOK, &info)
	if info["sso"] != true || info["name"] != "sam" {
		t.Fatalf("not offered with a licence key: %v", info)
	}
	var start obj
	visitor.json(http.MethodPost, "/api/auth/link/sso", obj{"token": token}, http.StatusOK, &start)
	visitor.http.CheckRedirect = nil
	res, err := visitor.http.Get(start["redirect"].(string))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	var st obj
	visitor.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
	if st["authenticated"] != true || st["role"] != "member" || st["username"] != "sam" {
		t.Fatalf("not signed in as the invited member (ended at %s): %v", res.Request.URL, st)
	}
	visitor.csrf, _ = st["csrfToken"].(string)
	var links []obj
	visitor.json(http.MethodGet, "/api/auth/identities", nil, http.StatusOK, &links)
	if len(links) != 1 || links[0]["onlySignIn"] != true || links[0]["email"] != "sam@outlook.com" {
		t.Fatalf("links: %v", links)
	}
	visitor.do(http.MethodDelete, "/api/auth/identities/sam-sub", nil, http.StatusConflict)

	// The invitation is used up.
	visitor.do(http.MethodPost, "/api/auth/link/inspect", obj{"token": token}, http.StatusUnauthorized)
}

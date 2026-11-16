//go:build ee

package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"trisse-home/ee/licence/domain"
	"trisse-home/internal/platform/config"
	"trisse-home/internal/platform/database"
	"trisse-home/internal/platform/oauth"
)

// TestConnectedAccounts runs #326 in the ee build: an admin registers a
// service, a member connects through a fake OAuth provider, disconnects
// (the token is revoked), and a deleted member's tokens are gone.
func TestConnectedAccounts(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved, savedMembers, savedEndpoints := licencePublicKey, membersEnabled, connectionEndpoints
	t.Cleanup(func() { licencePublicKey, membersEnabled, connectionEndpoints = saved, savedMembers, savedEndpoints })

	var mu sync.Mutex
	revoked := map[string]bool{}
	provider := http.NewServeMux()
	provider.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("code") != "good" || r.Form.Get("code_verifier") == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "acc", "refresh_token": "ref", "expires_in": 3600, "token_type": "Bearer"})
	})
	provider.HandleFunc("POST /revoke", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		mu.Lock()
		revoked[r.Form.Get("token")] = true
		mu.Unlock()
	})
	idp := httptest.NewServer(provider)
	t.Cleanup(idp.Close)

	licencePublicKey, membersEnabled = pub, true
	connectionEndpoints = func(string) (oauth.Endpoint, bool) {
		return oauth.Endpoint{AuthURL: idp.URL + "/auth", TokenURL: idp.URL + "/token", RevokeURL: idp.URL + "/revoke", PKCE: true}, true
	}

	up := fakeUpstream(t)
	var dbPath string
	srv := newTestServer(t, up.URL, func(c *config.Config) {
		c.PublicURL, _ = url.Parse("https://home.example")
		dbPath = c.DatabasePath
	})
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"connected-accounts"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	var m obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &m)
	anna := newClient(t, srv.URL)
	var as obj
	anna.json(http.MethodPost, "/api/auth/login", obj{"username": "anna", "password": "member-password-1"}, http.StatusOK, &as)
	anna.csrf = as["csrfToken"].(string)

	anna.do(http.MethodPost, "/api/connections/notion", nil, http.StatusConflict) // no client registered yet
	anna.do(http.MethodPut, "/api/connections/clients/notion", obj{"clientId": "x"}, http.StatusForbidden)
	owner.do(http.MethodPut, "/api/connections/clients/notion", obj{"clientId": "cid", "clientSecret": "secret"}, http.StatusNoContent)
	var clients struct {
		RedirectURL string `json:"redirectUrl"`
		Clients     []obj  `json:"clients"`
	}
	owner.json(http.MethodGet, "/api/connections/clients", nil, http.StatusOK, &clients)
	if clients.RedirectURL != "https://home.example/api/connections/callback" {
		t.Fatalf("redirect URL %q", clients.RedirectURL)
	}
	for _, c := range clients.Clients {
		if _, leaked := c["clientSecret"]; leaked {
			t.Fatalf("secret in the answer: %v", c)
		}
	}

	var start struct {
		URL string `json:"url"`
	}
	anna.json(http.MethodPost, "/api/connections/notion", nil, http.StatusOK, &start)
	u, _ := url.Parse(start.URL)
	state := u.Query().Get("state")
	if u.Query().Get("code_challenge") == "" || state == "" {
		t.Fatalf("auth url without PKCE or state: %s", start.URL)
	}
	// The provider sends anna's browser back with a code.
	anna.http.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	res, err := anna.http.Get(srv.URL + "/api/connections/callback?code=good&state=" + url.QueryEscape(state))
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if loc := res.Header.Get("Location"); res.StatusCode != http.StatusSeeOther || loc != "/?connected=notion&settings=connections" {
		t.Fatalf("callback: %d %s", res.StatusCode, loc)
	}
	var list []obj
	anna.json(http.MethodGet, "/api/connections", nil, http.StatusOK, &list)
	connected := false
	for _, c := range list {
		if c["service"] == "notion" && c["status"] == "connected" {
			connected = true
		}
		if _, leaked := c["token"]; leaked {
			t.Fatalf("token in the answer: %v", c)
		}
	}
	if !connected {
		t.Fatalf("not connected: %v", list)
	}

	anna.do(http.MethodDelete, "/api/connections/notion", nil, http.StatusNoContent)
	mu.Lock()
	if !revoked["ref"] {
		t.Fatal("token not revoked at the service")
	}
	mu.Unlock()

	// Connect again, then delete the member: the tokens go with the account.
	anna.json(http.MethodPost, "/api/connections/notion", nil, http.StatusOK, &start)
	u, _ = url.Parse(start.URL)
	res, _ = anna.http.Get(srv.URL + "/api/connections/callback?code=good&state=" + url.QueryEscape(u.Query().Get("state")))
	_ = res.Body.Close()
	owner.do(http.MethodDelete, "/api/auth/members/"+m["id"].(string), nil, http.StatusNoContent)
	db, err := database.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM connection_tokens`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("tokens of the deleted member: %d %v", n, err)
	}
}

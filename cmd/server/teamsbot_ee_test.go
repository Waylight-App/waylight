//go:build ee

package main

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"waylight/ee/licence/domain"
	"waylight/internal/platform/config"
)

// TestTeamsRouteIsOptIn checks ADR 0023: without TEAMS_BOT_APP_ID the route
// does not exist; with it, a request without a valid Bot Connector token is
// refused before anything runs.
func TestTeamsRouteIsOptIn(t *testing.T) {
	up := fakeUpstream(t)
	post := func(base string) int {
		res, err := http.Post(base+"/api/teams/messages", "application/json",
			strings.NewReader(`{"type":"invoke","name":"composeExtension/query","serviceUrl":"https://x/","channelId":"msteams"}`))
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res.StatusCode
	}
	t.Setenv("TEAMS_BOT_APP_ID", "")
	if code := post(newTestServer(t, up.URL).URL); code == http.StatusOK || code == http.StatusForbidden {
		t.Fatalf("route exists without opt-in: %d", code)
	}
	saved := teamsBotMetadata
	teamsBotMetadata = up.URL + "/no-metadata-here"
	t.Cleanup(func() { teamsBotMetadata = saved })
	t.Setenv("TEAMS_BOT_APP_ID", "app-1")
	if code := post(newTestServer(t, up.URL).URL); code != http.StatusForbidden {
		t.Fatalf("unsigned request: %d, want 403", code)
	}

	// A correctly signed request passes every middleware and is answered.
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var connector *httptest.Server
	connector = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b64 := base64.RawURLEncoding.EncodeToString
		if r.URL.Path == "/meta" {
			_ = json.NewEncoder(w).Encode(map[string]string{"jwks_uri": connector.URL + "/keys"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{{"kty": "RSA", "kid": "k1",
			"n": b64(priv.N.Bytes()), "e": b64(big.NewInt(int64(priv.E)).Bytes()), "endorsements": []string{"msteams"}}}})
	}))
	t.Cleanup(connector.Close)
	teamsBotMetadata = connector.URL + "/meta"
	enc := func(v any) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
	signed := enc(map[string]string{"alg": "RS256", "kid": "k1"}) + "." +
		enc(map[string]any{"iss": "https://api.botframework.com", "aud": "app-1", "exp": time.Now().Add(time.Hour).Unix(), "serviceUrl": "https://x/"})
	digest := sha256.Sum256([]byte(signed))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
	srv := newTestServer(t, up.URL)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/teams/messages", strings.NewReader(
		`{"type":"invoke","name":"composeExtension/query","serviceUrl":"https://x/","channelId":"msteams","from":{"aadObjectId":"oid"},`+
			`"value":{"parameters":[{"name":"q","value":"vpn"}]}}`))
	req.Header.Set("Authorization", "Bearer "+signed+"."+base64.RawURLEncoding.EncodeToString(sig))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"composeExtension"`) {
		t.Fatalf("signed request: %d %s", res.StatusCode, body)
	}
}

// TestTeamsTabPage checks that the tab page can be framed by Teams through
// the whole server (the security middleware denies framing everywhere
// else) and only with the licence.
func TestTeamsTabPage(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved := licencePublicKey
	licencePublicKey = pub
	t.Cleanup(func() { licencePublicKey = saved })
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.PublicURL, _ = url.Parse("https://home.example") })
	get := func(path string) *http.Response {
		res, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res
	}
	if res := get("/teams/tab"); res.StatusCode != http.StatusNotFound {
		t.Fatalf("without licence: %d", res.StatusCode)
	}
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"chat-apps"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)
	res := get("/teams/tab")
	if res.StatusCode != http.StatusOK || res.Header.Get("X-Frame-Options") != "" ||
		!strings.Contains(res.Header.Get("Content-Security-Policy"), "frame-ancestors https://*.cloud.microsoft") {
		t.Fatalf("tab: %d %v", res.StatusCode, res.Header)
	}
	if res := get("/"); res.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatal("the app itself became frameable")
	}
}

package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"trisse-home/internal/platform/config"
)

// fakeMicrosoft is a device-flow provider (RFC 8628) like Microsoft's
// identity platform: every code is pending at the first poll and confirmed
// at the second.
func fakeMicrosoft(t *testing.T, subject, email string) *httptest.Server {
	t.Helper()
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	b64 := base64.RawURLEncoding.EncodeToString
	var srv *httptest.Server
	var mu sync.Mutex
	polls := map[string]int{}
	n := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"issuer": srv.URL, "authorization_endpoint": srv.URL + "/auth",
			"token_endpoint": srv.URL + "/token", "jwks_uri": srv.URL + "/jwks", "device_authorization_endpoint": srv.URL + "/devicecode"})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{
			{"kty": "RSA", "kid": "k1", "n": b64(key.N.Bytes()), "e": b64(big.NewInt(int64(key.E)).Bytes())}}})
	})
	mux.HandleFunc("/devicecode", func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		n++
		code := fmt.Sprintf("dev-%d", n)
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{"device_code": code, "user_code": "ABCD-1234",
			"verification_uri": srv.URL + "/link", "expires_in": 900, "interval": 1})
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		code := r.Form.Get("device_code")
		mu.Lock()
		polls[code]++
		first := polls[code] == 1
		mu.Unlock()
		if first {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "authorization_pending"})
			return
		}
		h, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1"})
		c, _ := json.Marshal(map[string]any{"iss": srv.URL, "sub": subject, "aud": "trisse", "email": email,
			"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()})
		input := b64(h) + "." + b64(c)
		digest := sha256.Sum256([]byte(input))
		sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		_ = json.NewEncoder(w).Encode(map[string]string{"id_token": input + "." + b64(sig)})
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// TestSignInWithMicrosoftByCode links and signs in with a Microsoft account
// by code (#535): no redirect address and no setup per server.
func TestSignInWithMicrosoftByCode(t *testing.T) {
	ms := fakeMicrosoft(t, "ms-sub", "me@outlook.com")
	app := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) {
		c.MicrosoftClientID, c.MicrosoftIssuer = "trisse", ms.URL
	})
	owner := newClient(t, app.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var st obj
	newClient(t, app.URL).json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
	if st["sso"].(map[string]any)["microsoft"] != true {
		t.Fatalf("status: %v", st)
	}

	// poll waits the interval and polls until the flow ends.
	poll := func(c *client, flow obj) (int, obj) {
		for range 10 {
			time.Sleep(1100 * time.Millisecond)
			var out obj
			code := postStatus(c, "/api/auth/microsoft/poll", obj{"flow": flow["flow"]}, &out)
			if code != http.StatusAccepted {
				return code, out
			}
		}
		t.Fatal("the flow never ended")
		return 0, nil
	}

	// Not linked yet: refused.
	b := newClient(t, app.URL)
	var flow obj
	b.json(http.MethodPost, "/api/auth/microsoft/start", nil, http.StatusOK, &flow)
	if flow["userCode"] != "ABCD-1234" || flow["verificationUri"] != ms.URL+"/link" {
		t.Fatalf("flow: %v", flow)
	}
	if code, out := poll(b, flow); code != http.StatusUnauthorized {
		t.Fatalf("unlinked account signed in: %d %v", code, out)
	}

	// Link from the settings.
	owner.json(http.MethodPost, "/api/auth/microsoft/link", nil, http.StatusOK, &flow)
	if code, out := poll(owner, flow); code != http.StatusOK || out["linked"] != true {
		t.Fatalf("link: %d %v", code, out)
	}
	var links []obj
	owner.json(http.MethodGet, "/api/auth/identities", nil, http.StatusOK, &links)
	if len(links) != 1 || links[0]["email"] != "me@outlook.com" {
		t.Fatalf("links: %v", links)
	}

	// Now the code signs in.
	b = newClient(t, app.URL)
	b.json(http.MethodPost, "/api/auth/microsoft/start", nil, http.StatusOK, &flow)
	if code, out := poll(b, flow); code != http.StatusOK || out["csrfToken"] == nil {
		t.Fatalf("sign-in: %d %v", code, out)
	}
	b.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
	if st["authenticated"] != true || st["role"] != "owner" {
		t.Fatalf("status after sign-in: %v", st)
	}

	// A used or unknown flow is gone.
	b.do(http.MethodPost, "/api/auth/microsoft/poll", obj{"flow": flow["flow"]}, http.StatusConflict)
}

// postStatus posts like client.do but returns the status instead of
// expecting one, for polls that answer 202 until they are done.
func postStatus(c *client, path string, body, dst any) int {
	c.t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	}
	res, err := c.http.Do(req)
	if err != nil {
		c.t.Fatalf("POST %s: %v", path, err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	_ = json.Unmarshal(out, dst)
	return res.StatusCode
}

package oauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestAuthCodeURL(t *testing.T) {
	cfg := Config{Endpoint: Endpoint{AuthURL: "https://login.example/authorize", Scopes: []string{"offline_access", "Tasks.ReadWrite"},
		ExtraAuth: url.Values{"prompt": {"consent"}}, PKCE: true}, ClientID: "app", RedirectURL: "https://home.example/cb"}
	u, _ := url.Parse(AuthCodeURL(cfg, "st", "verifier"))
	q := u.Query()
	if q.Get("client_id") != "app" || q.Get("state") != "st" || q.Get("scope") != "offline_access Tasks.ReadWrite" || q.Get("prompt") != "consent" ||
		q.Get("code_challenge") != Challenge("verifier") || q.Get("code_challenge_method") != "S256" || q.Get("redirect_uri") != "https://home.example/cb" {
		t.Fatal(u)
	}
	cfg.PKCE = false
	if strings.Contains(AuthCodeURL(cfg, "st", "v"), "code_challenge") {
		t.Fatal("PKCE sent although off")
	}
	if a, b := NewVerifier(), NewVerifier(); a == b || len(a) < 40 {
		t.Fatal("verifier")
	}
}

func TestTokenFlow(t *testing.T) {
	var got []url.Values
	var basic string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		got = append(got, r.PostForm)
		if u, p, ok := r.BasicAuth(); ok {
			basic = u + ":" + p
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/revoke":
			w.WriteHeader(http.StatusOK)
		case r.PostForm.Get("refresh_token") == "dead":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
		case r.PostForm.Get("grant_type") == "refresh_token":
			_, _ = w.Write([]byte(`{"access_token":"a2","expires_in":3600,"token_type":"Bearer"}`))
		default:
			_, _ = w.Write([]byte(`{"access_token":"a1","refresh_token":"r1","expires_in":3600,"token_type":"Bearer"}`))
		}
	}))
	defer srv.Close()
	now := time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
	c := Client{HTTP: srv.Client(), Now: func() time.Time { return now }}
	cfg := Config{Endpoint: Endpoint{TokenURL: srv.URL + "/token", RevokeURL: srv.URL + "/revoke", PKCE: true}, ClientID: "app", ClientSecret: "s3cret", RedirectURL: "https://home.example/cb"}
	ctx := context.Background()

	tok, err := c.Exchange(ctx, cfg, "code1", "ver")
	if err != nil || tok.Access != "a1" || tok.Refresh != "r1" || !tok.Expiry.Equal(now.Add(time.Hour)) {
		t.Fatalf("%+v %v", tok, err)
	}
	if f := got[0]; f.Get("code") != "code1" || f.Get("code_verifier") != "ver" || f.Get("client_secret") != "s3cret" || f.Get("redirect_uri") != cfg.RedirectURL {
		t.Fatal(f)
	}
	// A refresh without a new refresh token keeps the old one.
	if tok, err = c.Refresh(ctx, cfg, "r1"); err != nil || tok.Access != "a2" || tok.Refresh != "r1" {
		t.Fatalf("%+v %v", tok, err)
	}
	if _, err := c.Refresh(ctx, cfg, "dead"); !errors.Is(err, ErrRejected) {
		t.Fatal(err)
	}
	cfg.BasicAuth = true
	if _, err := c.Exchange(ctx, cfg, "code2", "ver"); err != nil || basic != "app:s3cret" || got[len(got)-1].Get("client_secret") != "" {
		t.Fatalf("basic auth: %q %v", basic, got[len(got)-1])
	}
	if err := c.Revoke(ctx, cfg, "a2"); err != nil || got[len(got)-1].Get("token") != "a2" {
		t.Fatal(err)
	}
	cfg.RevokeURL = ""
	if err := c.Revoke(ctx, cfg, "a2"); err != nil {
		t.Fatal(err)
	}
}

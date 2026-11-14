// Package oauth is a small OAuth 2.0 authorization-code client (RFC 6749)
// with PKCE (RFC 7636), refresh and revocation, for integrations that sign
// in with the user's own app registration (to-do providers, #171). It is a
// framework detail: modules receive it through their own ports.
package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Endpoint describes a provider.
type Endpoint struct {
	AuthURL   string
	TokenURL  string
	RevokeURL string // "" = the provider has no revocation endpoint
	Scopes    []string
	// ExtraAuth are added to the authorization URL (e.g. access_type=offline).
	ExtraAuth url.Values
	// BasicAuth sends the client credentials as HTTP Basic instead of form fields.
	BasicAuth bool
	// PKCE adds code_challenge (S256); providers that do not know it ignore it,
	// but some reject it, so it is per provider.
	PKCE bool
}

// Config is an endpoint with the user's app registration.
type Config struct {
	Endpoint
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Token is what the token endpoint returned.
type Token struct {
	Access  string
	Refresh string
	Expiry  time.Time // zero = unknown
}

// ErrRejected is returned when the provider refuses a grant (expired or
// revoked refresh token, wrong client secret); the user has to sign in again.
var ErrRejected = errors.New("oauth: the provider rejected the grant")

// Client talks to token endpoints.
type Client struct {
	HTTP *http.Client
	Now  func() time.Time
}

// NewVerifier returns a random PKCE code verifier (also usable as state).
func NewVerifier() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Challenge is the S256 code challenge of a verifier.
func Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// AuthCodeURL is where the browser signs in.
func AuthCodeURL(cfg Config, state, verifier string) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {cfg.RedirectURL},
		"state":         {state},
	}
	if len(cfg.Scopes) > 0 {
		q.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	if cfg.PKCE {
		q.Set("code_challenge", Challenge(verifier))
		q.Set("code_challenge_method", "S256")
	}
	for k, v := range cfg.ExtraAuth {
		q[k] = v
	}
	sep := "?"
	if strings.Contains(cfg.AuthURL, "?") {
		sep = "&"
	}
	return cfg.AuthURL + sep + q.Encode()
}

// Exchange trades an authorization code for tokens.
func (c Client) Exchange(ctx context.Context, cfg Config, code, verifier string) (Token, error) {
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {cfg.RedirectURL}}
	if cfg.PKCE {
		form.Set("code_verifier", verifier)
	}
	return c.token(ctx, cfg, form)
}

// Refresh gets a new access token; a provider that sends no new refresh
// token keeps the old one valid, so it is carried over.
func (c Client) Refresh(ctx context.Context, cfg Config, refresh string) (Token, error) {
	t, err := c.token(ctx, cfg, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refresh}})
	if err == nil && t.Refresh == "" {
		t.Refresh = refresh
	}
	return t, err
}

func (c Client) token(ctx context.Context, cfg Config, form url.Values) (Token, error) {
	if !cfg.BasicAuth {
		form.Set("client_id", cfg.ClientID)
		if cfg.ClientSecret != "" {
			form.Set("client_secret", cfg.ClientSecret)
		}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if cfg.BasicAuth {
		req.SetBasicAuth(url.QueryEscape(cfg.ClientID), url.QueryEscape(cfg.ClientSecret))
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("token request failed: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return Token{}, fmt.Errorf("read token answer: %w", err)
	}
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int64  `json:"expires_in"`
		Error        string `json:"error"`
	}
	_ = json.Unmarshal(raw, &body)
	switch {
	case res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusUnauthorized || body.Error == "invalid_grant" || body.Error == "invalid_client":
		return Token{}, fmt.Errorf("%w (%s)", ErrRejected, cmpOr(body.Error, res.Status))
	case res.StatusCode >= 300:
		return Token{}, fmt.Errorf("token endpoint answered %s", res.Status)
	case body.AccessToken == "":
		return Token{}, errors.New("oauth: the token answer has no access token")
	}
	t := Token{Access: body.AccessToken, Refresh: body.RefreshToken}
	if body.ExpiresIn > 0 {
		now := time.Now
		if c.Now != nil {
			now = c.Now
		}
		t.Expiry = now().Add(time.Duration(body.ExpiresIn) * time.Second)
	}
	return t, nil
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// Revoke invalidates a token at the provider (RFC 7009); providers without
// a revocation endpoint are skipped.
func (c Client) Revoke(ctx context.Context, cfg Config, token string) error {
	if cfg.RevokeURL == "" || token == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.RevokeURL, strings.NewReader(url.Values{"token": {token}}.Encode()))
	if err != nil {
		return fmt.Errorf("build revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 && res.StatusCode != http.StatusBadRequest { // 400 = already invalid
		return fmt.Errorf("revoke endpoint answered %s", res.Status)
	}
	return nil
}

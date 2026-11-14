package todos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/modules/todos/usecase"
	"trisse-home/internal/platform/oauth"
)

// CallbackPath is where the providers send the browser back after signing
// in; users register PUBLIC_URL + CallbackPath in their app registration.
const CallbackPath = "/api/todos/settings/oauth/callback"

// endpoints of the providers that are connected by signing in (#171).
var endpoints = map[domain.SourceKind]oauth.Endpoint{
	domain.KindMicrosoft: { //nolint:gosec // public endpoint URLs, no credentials
		AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		Scopes:   []string{"offline_access", "Tasks.ReadWrite"},
		PKCE:     true, // Microsoft has no revocation endpoint
	},
	domain.KindGoogle: { //nolint:gosec // public endpoint URLs, no credentials
		AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:  "https://oauth2.googleapis.com/token",
		RevokeURL: "https://oauth2.googleapis.com/revoke",
		Scopes:    []string{"https://www.googleapis.com/auth/tasks"},
		// offline + consent: Google sends a refresh token every time.
		ExtraAuth: url.Values{"access_type": {"offline"}, "prompt": {"consent"}},
		PKCE:      true,
	},
	domain.KindTickTick: { //nolint:gosec // public endpoint URLs, no credentials
		AuthURL:   "https://ticktick.com/oauth/authorize",
		TokenURL:  "https://ticktick.com/oauth/token",
		Scopes:    []string{"tasks:read", "tasks:write"},
		BasicAuth: true, // TickTick wants the client credentials as HTTP Basic
	},
}

// oauthPort implements usecase.OAuth on the platform's OAuth client.
type oauthPort struct {
	client   oauth.Client
	redirect string
}

func newOAuthPort(client *http.Client, publicURL *url.URL, now func() time.Time) *oauthPort {
	p := &oauthPort{client: oauth.Client{HTTP: client, Now: now}}
	if publicURL != nil {
		p.redirect = publicURL.Scheme + "://" + publicURL.Host + CallbackPath
	}
	return p
}

func (p *oauthPort) config(kind domain.SourceKind, clientID, secret string) (oauth.Config, error) {
	e, ok := endpoints[kind]
	if !ok {
		return oauth.Config{}, fmt.Errorf("no OAuth endpoint for %q", kind)
	}
	return oauth.Config{Endpoint: e, ClientID: clientID, ClientSecret: secret, RedirectURL: p.redirect}, nil
}

func (p *oauthPort) RedirectURL() string { return p.redirect }

func (p *oauthPort) NewState() string { return oauth.NewVerifier() }

func (p *oauthPort) AuthURL(kind domain.SourceKind, clientID, state, verifier string) string {
	cfg, err := p.config(kind, clientID, "")
	if err != nil {
		return ""
	}
	return oauth.AuthCodeURL(cfg, state, verifier)
}

func grant(t oauth.Token, secret string) domain.OAuthGrant {
	return domain.OAuthGrant{ClientSecret: secret, Access: t.Access, Refresh: t.Refresh, Expiry: t.Expiry}
}

func (p *oauthPort) Exchange(ctx context.Context, kind domain.SourceKind, clientID, secret, code, verifier string) (domain.OAuthGrant, error) {
	cfg, err := p.config(kind, clientID, secret)
	if err != nil {
		return domain.OAuthGrant{}, err
	}
	t, err := p.client.Exchange(ctx, cfg, code, verifier)
	return grant(t, secret), err
}

func (p *oauthPort) Refresh(ctx context.Context, kind domain.SourceKind, clientID string, g domain.OAuthGrant) (domain.OAuthGrant, error) {
	cfg, err := p.config(kind, clientID, g.ClientSecret)
	if err != nil {
		return domain.OAuthGrant{}, err
	}
	t, err := p.client.Refresh(ctx, cfg, g.Refresh)
	if errors.Is(err, oauth.ErrRejected) {
		return domain.OAuthGrant{}, fmt.Errorf("%w: %w", usecase.ErrGrantRejected, err)
	}
	return grant(t, g.ClientSecret), err
}

func (p *oauthPort) Revoke(ctx context.Context, kind domain.SourceKind, clientID string, g domain.OAuthGrant) error {
	cfg, err := p.config(kind, clientID, g.ClientSecret)
	if err != nil {
		return err
	}
	token := g.Refresh
	if token == "" {
		token = g.Access
	}
	return p.client.Revoke(ctx, cfg, token)
}

package calendar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"waylight/internal/modules/calendar/domain"
	"waylight/internal/modules/calendar/usecase"
	"waylight/internal/platform/oauth"
)

// CallbackPath is where the providers send the browser back after signing
// in; users register PUBLIC_URL + CallbackPath in their app registration.
const CallbackPath = "/api/calendar/accounts/oauth/callback"

// endpoints of the calendar providers (#282). Scopes allow reading and
// creating events; the to-do lists (#171) use their own sign-in.
var endpoints = map[domain.Provider]oauth.Endpoint{
	domain.ProviderGoogle: { //nolint:gosec // public endpoint URLs, no credentials
		AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:  "https://oauth2.googleapis.com/token",
		RevokeURL: "https://oauth2.googleapis.com/revoke",
		Scopes:    []string{"https://www.googleapis.com/auth/calendar"},
		// offline + consent: Google sends a refresh token every time.
		ExtraAuth: url.Values{"access_type": {"offline"}, "prompt": {"consent"}},
		PKCE:      true,
	},
	domain.ProviderMicrosoft: { //nolint:gosec // public endpoint URLs, no credentials
		AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		Scopes:   []string{"offline_access", "User.Read", "Calendars.ReadWrite"},
		PKCE:     true, // Microsoft has no revocation endpoint
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

func (p *oauthPort) config(provider domain.Provider, clientID, secret string) (oauth.Config, error) {
	e, ok := endpoints[provider]
	if !ok {
		return oauth.Config{}, fmt.Errorf("no OAuth endpoint for %q", provider)
	}
	return oauth.Config{Endpoint: e, ClientID: clientID, ClientSecret: secret, RedirectURL: p.redirect}, nil
}

func (p *oauthPort) RedirectURL() string { return p.redirect }

func (p *oauthPort) NewState() string { return oauth.NewVerifier() }

func (p *oauthPort) AuthURL(provider domain.Provider, clientID, state, verifier string) string {
	cfg, err := p.config(provider, clientID, "")
	if err != nil {
		return ""
	}
	return oauth.AuthCodeURL(cfg, state, verifier)
}

func grant(t oauth.Token, secret string) domain.Grant {
	return domain.Grant{ClientSecret: secret, Access: t.Access, Refresh: t.Refresh, Expiry: t.Expiry}
}

func (p *oauthPort) Exchange(ctx context.Context, provider domain.Provider, clientID, secret, code, verifier string) (domain.Grant, error) {
	cfg, err := p.config(provider, clientID, secret)
	if err != nil {
		return domain.Grant{}, err
	}
	t, err := p.client.Exchange(ctx, cfg, code, verifier)
	return grant(t, secret), err
}

func (p *oauthPort) Refresh(ctx context.Context, provider domain.Provider, clientID string, g domain.Grant) (domain.Grant, error) {
	cfg, err := p.config(provider, clientID, g.ClientSecret)
	if err != nil {
		return domain.Grant{}, err
	}
	t, err := p.client.Refresh(ctx, cfg, g.Refresh)
	if errors.Is(err, oauth.ErrRejected) {
		return domain.Grant{}, fmt.Errorf("%w: %w", usecase.ErrGrantRejected, err)
	}
	return grant(t, g.ClientSecret), err
}

func (p *oauthPort) Revoke(ctx context.Context, provider domain.Provider, clientID string, g domain.Grant) error {
	cfg, err := p.config(provider, clientID, g.ClientSecret)
	if err != nil {
		return err
	}
	token := g.Refresh
	if token == "" {
		token = g.Access
	}
	return p.client.Revoke(ctx, cfg, token)
}

package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"trisse-home/ee/connections/domain"
	"trisse-home/ee/connections/usecase"
	"trisse-home/internal/platform/oauth"
)

// CallbackPath is where services send the browser back.
const CallbackPath = "/api/connections/callback"

// service is one known service: its endpoints and the scopes the team
// integrations need (#327–#331).
type service struct {
	info     usecase.ServiceInfo
	endpoint func(c domain.Client) oauth.Endpoint
}

var services = []service{
	{
		info: usecase.ServiceInfo{ID: "microsoft", Name: "Microsoft 365"},
		endpoint: func(c domain.Client) oauth.Endpoint {
			tenant := c.Tenant
			if tenant == "" {
				tenant = "organizations"
			}
			base := "https://login.microsoftonline.com/" + url.PathEscape(tenant) + "/oauth2/v2.0/"
			return oauth.Endpoint{AuthURL: base + "authorize", TokenURL: base + "token", PKCE: true,
				Scopes: []string{"offline_access", "User.Read", "Sites.Read.All", "Files.Read.All", "ChannelMessage.Read.All", "Chat.Read"}}
		},
	},
	{
		info: usecase.ServiceInfo{ID: "atlassian", Name: "Confluence (Atlassian Cloud)"},
		endpoint: func(domain.Client) oauth.Endpoint {
			return oauth.Endpoint{AuthURL: "https://auth.atlassian.com/authorize", TokenURL: "https://auth.atlassian.com/oauth/token", //nolint:gosec // endpoint URLs, no credentials
				Scopes:    []string{"read:confluence-content.all", "search:confluence", "read:confluence-space.summary", "offline_access"},
				ExtraAuth: url.Values{"audience": {"api.atlassian.com"}, "prompt": {"consent"}}, PKCE: true}
		},
	},
	{
		info: usecase.ServiceInfo{ID: "notion", Name: "Notion"},
		endpoint: func(domain.Client) oauth.Endpoint {
			// Notion issues long-lived tokens without refresh and wants HTTP Basic.
			return oauth.Endpoint{AuthURL: "https://api.notion.com/v1/oauth/authorize", TokenURL: "https://api.notion.com/v1/oauth/token", //nolint:gosec // endpoint URLs, no credentials
				ExtraAuth: url.Values{"owner": {"user"}}, BasicAuth: true}
		},
	},
	{
		info: usecase.ServiceInfo{ID: "slack", Name: "Slack"},
		endpoint: func(domain.Client) oauth.Endpoint {
			// A user token (#330): Slack's search then follows the person's own
			// access. The rights are asked as user scopes, and the token comes
			// back under authed_user. The workspace's internal app must list the
			// same user scopes (docs/connected-accounts.md).
			return oauth.Endpoint{AuthURL: "https://slack.com/oauth/v2/authorize", TokenURL: "https://slack.com/api/oauth.v2.access", //nolint:gosec // endpoint URLs, no credentials
				RevokeURL:  "https://slack.com/api/auth.revoke",
				Scopes:     []string{"search:read.public", "search:read.private", "search:read.mpim", "search:read.im"},
				ScopeParam: "user_scope", UserToken: true}
		},
	},
}

// OAuth implements usecase.OAuth with the platform's OAuth client.
type OAuth struct {
	client   oauth.Client
	redirect string
	override func(id string) (oauth.Endpoint, bool) // tests point services at a fake provider
}

// NewOAuth returns the adapter; publicURL is PUBLIC_URL (nil disables
// connecting, because the service cannot send the browser back).
func NewOAuth(client *http.Client, now func() time.Time, publicURL *url.URL) *OAuth {
	o := &OAuth{client: oauth.Client{HTTP: client, Now: now}}
	if publicURL != nil {
		o.redirect = publicURL.Scheme + "://" + publicURL.Host + strings.TrimSuffix(publicURL.Path, "/") + CallbackPath
	}
	return o
}

// WithEndpoints replaces the endpoints of services, for tests.
func (o *OAuth) WithEndpoints(f func(id string) (oauth.Endpoint, bool)) *OAuth {
	o.override = f
	return o
}

// Services implements usecase.OAuth.
func (o *OAuth) Services() []usecase.ServiceInfo {
	out := make([]usecase.ServiceInfo, 0, len(services))
	for _, s := range services {
		out = append(out, s.info)
	}
	return out
}

// RedirectURL implements usecase.OAuth.
func (o *OAuth) RedirectURL() string { return o.redirect }

func (o *OAuth) config(c domain.Client) (oauth.Config, error) {
	for _, s := range services {
		if s.info.ID != c.Service {
			continue
		}
		e := s.endpoint(c)
		if o.override != nil {
			if fake, ok := o.override(c.Service); ok {
				e = fake
			}
		}
		return oauth.Config{Endpoint: e, ClientID: c.ClientID, ClientSecret: c.Secret, RedirectURL: o.redirect}, nil
	}
	return oauth.Config{}, fmt.Errorf("unknown service %q", c.Service)
}

// AuthCodeURL implements usecase.OAuth.
func (o *OAuth) AuthCodeURL(c domain.Client, state, verifier string) (string, error) {
	cfg, err := o.config(c)
	if err != nil {
		return "", err
	}
	return oauth.AuthCodeURL(cfg, state, verifier), nil
}

func token(t oauth.Token) domain.Token {
	return domain.Token{Access: t.Access, Refresh: t.Refresh, Expiry: t.Expiry}
}

// Exchange implements usecase.OAuth.
func (o *OAuth) Exchange(ctx context.Context, c domain.Client, code, verifier string) (domain.Token, error) {
	cfg, err := o.config(c)
	if err != nil {
		return domain.Token{}, err
	}
	t, err := o.client.Exchange(ctx, cfg, code, verifier)
	return token(t), err
}

// Refresh implements usecase.OAuth.
func (o *OAuth) Refresh(ctx context.Context, c domain.Client, refresh string) (domain.Token, error) {
	cfg, err := o.config(c)
	if err != nil {
		return domain.Token{}, err
	}
	t, err := o.client.Refresh(ctx, cfg, refresh)
	if errors.Is(err, oauth.ErrRejected) {
		return domain.Token{}, fmt.Errorf("%w: %w", usecase.ErrRejected, err)
	}
	return token(t), err
}

// Revoke implements usecase.OAuth.
func (o *OAuth) Revoke(ctx context.Context, c domain.Client, tok string) error {
	cfg, err := o.config(c)
	if err != nil {
		return err
	}
	return o.client.Revoke(ctx, cfg, tok)
}

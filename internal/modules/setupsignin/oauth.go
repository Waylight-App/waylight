package setupsignin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"waylight/internal/platform/oauth"
	"waylight/internal/shared/kernel"
)

// CallbackPath is where Home Assistant sends the browser back. It lies on
// the same origin as the client ID, so Home Assistant accepts it without
// fetching the client's page (IndieAuth redirect rules).
const CallbackPath = "/api/setup/signin/oauth/callback"

// homeAssistant implements usecase.HomeAssistantAuth on the platform's
// OAuth client.
type homeAssistant struct {
	client   oauth.Client
	clientID string
	redirect string
}

func newHomeAssistant(client *http.Client, publicURL *url.URL, now func() time.Time) *homeAssistant {
	h := &homeAssistant{client: oauth.Client{HTTP: client, Now: now}}
	if publicURL != nil {
		origin := publicURL.Scheme + "://" + publicURL.Host
		h.clientID, h.redirect = origin+"/", origin+CallbackPath
	}
	return h
}

func (h *homeAssistant) ClientID() string { return h.clientID }

func (h *homeAssistant) config(base string) oauth.Config {
	return oauth.Config{
		Endpoint: oauth.Endpoint{AuthURL: base + "/auth/authorize", TokenURL: base + "/auth/token"},
		ClientID: h.clientID, RedirectURL: h.redirect,
	}
}

func (h *homeAssistant) AuthURL(base, state string) string {
	return oauth.AuthCodeURL(h.config(base), state, "")
}

func (h *homeAssistant) Exchange(ctx context.Context, base, code string) (kernel.SignInGrant, error) {
	t, err := h.client.Exchange(ctx, h.config(base), code, "")
	if errors.Is(err, oauth.ErrRejected) {
		return kernel.SignInGrant{}, fmt.Errorf("%w: %w", kernel.ErrTargetAuth, err)
	}
	if err != nil {
		return kernel.SignInGrant{}, err
	}
	if t.Refresh == "" {
		return kernel.SignInGrant{}, fmt.Errorf("%w: the token answer has no refresh token", kernel.ErrTargetWrongService)
	}
	return kernel.SignInGrant{Access: t.Access, Expiry: t.Expiry, Secret: t.Refresh, ClientID: h.clientID}, nil
}

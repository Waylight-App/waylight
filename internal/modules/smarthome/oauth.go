package smarthome

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/platform/oauth"
	"waylight/internal/shared/kernel"
)

// renewer implements usecase.TokenRenewer with Home Assistant's token
// endpoint (#654). Home Assistant sends no new refresh token; the client
// keeps the old one.
type renewer struct{ client oauth.Client }

func newRenewer(client *http.Client, now func() time.Time) renewer {
	return renewer{client: oauth.Client{HTTP: client, Now: now}}
}

func (r renewer) Renew(ctx context.Context, s domain.Settings) (domain.Settings, error) {
	cfg := oauth.Config{Endpoint: oauth.Endpoint{TokenURL: s.URL + "/auth/token"}, ClientID: s.ClientID}
	t, err := r.client.Refresh(ctx, cfg, s.Refresh)
	if errors.Is(err, oauth.ErrRejected) {
		return domain.Settings{}, fmt.Errorf("%w: %w", kernel.ErrTargetAuth, err)
	}
	if err != nil {
		return domain.Settings{}, err
	}
	s.Token, s.Refresh, s.Expiry = t.Access, t.Refresh, t.Expiry
	return s, nil
}

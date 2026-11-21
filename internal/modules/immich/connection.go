package immich

import (
	"context"
	"errors"

	"waylight/internal/modules"
	"waylight/internal/modules/immich/domain"
	"waylight/internal/modules/immich/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Immich connection (#610); the check is the
// statistics call SaveSettings makes.
func connectionSpecs(client usecase.Client) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "immich", Title: "Immich",
		Fields: []modules.CredentialField{{Name: "apiKey", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["apiKey"])
			if err != nil {
				return err
			}
			if _, err := client.Stats(ctx, s); err != nil {
				return kernel.SetupFailure("Immich", err)
			}
			return nil
		},
	}}
}

// savedConnections returns the module settings as a connection to copy.
func savedConnections(ctx context.Context, repo usecase.SettingsRepository) ([]modules.SavedConnection, error) {
	s, err := repo.Load(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []modules.SavedConnection{{Source: "immich", Service: "immich", URL: s.URL, Credential: map[string]string{"apiKey": s.APIKey}}}, nil
}

// connections adapts the integrations resolver to usecase.Connections.
type connections struct{ r modules.ConnectionResolver }

// Settings implements usecase.Connections.
func (c connections) Settings(ctx context.Context, id string) (domain.Settings, error) {
	rc, err := c.r.ResolveConnection(ctx, "immich", id)
	if err != nil {
		return domain.Settings{}, err
	}
	return domain.NewSettings(rc.URL, rc.Credential["apiKey"])
}

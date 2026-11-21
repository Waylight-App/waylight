package immich

import (
	"context"

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

package grafana

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/grafana/domain"
	"waylight/internal/modules/grafana/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Grafana connection (#610); the check is the
// search SaveSettings makes.
func connectionSpecs(client usecase.Client) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "grafana", Title: "Grafana",
		Fields: []modules.CredentialField{{Name: "token", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["token"])
			if err != nil {
				return err
			}
			if _, err := client.Search(ctx, s, ""); err != nil {
				return kernel.SetupFailure("Grafana", err)
			}
			return nil
		},
	}}
}

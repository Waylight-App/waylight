package nextcloud

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/nextcloud/domain"
	"waylight/internal/modules/nextcloud/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Nextcloud connection (#610); the check is
// the status call SaveSettings makes.
func connectionSpecs(source usecase.StatusSource) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "nextcloud", Title: "Nextcloud",
		Fields: []modules.CredentialField{{Name: "username"}, {Name: "appPassword", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["username"], cred["appPassword"])
			if err != nil {
				return err
			}
			if _, err := source.Status(ctx, s); err != nil {
				return kernel.SetupFailure("Nextcloud", err)
			}
			return nil
		},
	}}
}

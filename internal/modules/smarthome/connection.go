package smarthome

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/modules/smarthome/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Home Assistant connection (#610); the check
// reads the states like SaveSettings.
func connectionSpecs(provider usecase.Provider) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "homeassistant", Title: "Home Assistant",
		Fields: []modules.CredentialField{{Name: "token", Secret: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["token"])
			if err != nil {
				return err
			}
			if _, err := provider.States(ctx, s); err != nil {
				return kernel.SetupFailure("Home Assistant", err)
			}
			return nil
		},
	}}
}

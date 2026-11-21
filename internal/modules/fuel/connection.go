package fuel

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/fuel/domain"
	"waylight/internal/modules/fuel/usecase"
	"waylight/internal/shared/kernel"
)

// checkLat and checkLon are a point the check asks prices around:
// Tankerkönig has no request that tests only the key, and the place of the
// widget is a setting of the module, not of the connection.
const checkLat, checkLon = 52.52, 13.405

// connectionSpecs declares the Tankerkönig connection (#610); the address
// is fixed, only the API key is entered.
func connectionSpecs(provider usecase.FuelPriceProvider, apiURL string) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "tankerkoenig", Title: "Tankerkönig", FixedURL: apiURL,
		Fields: []modules.CredentialField{{Name: "apiKey", Secret: true}},
		Check: func(ctx context.Context, _ string, cred map[string]string) error {
			s, err := domain.NewSettings(cred["apiKey"], checkLat, checkLon, "", domain.MinRadiusKm, string(domain.FuelE5))
			if err != nil {
				return err
			}
			if _, err := provider.Around(ctx, s); err != nil {
				return kernel.SetupFailure("Tankerkönig", err)
			}
			return nil
		},
	}}
}

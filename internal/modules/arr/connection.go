package arr

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/arr/domain"
	"waylight/internal/modules/arr/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares Sonarr, Radarr and Prowlarr as services of their
// own (#610); the check is the one AddInstance makes.
func connectionSpecs(providers map[domain.App]usecase.Provider) []modules.ConnectionSpec {
	out := make([]modules.ConnectionSpec, 0, len(providers))
	for _, app := range []domain.App{domain.AppSonarr, domain.AppRadarr, domain.AppProwlarr} {
		provider := providers[app]
		out = append(out, modules.ConnectionSpec{
			Service: string(app), Title: app.Title(),
			Fields: []modules.CredentialField{{Name: "apiKey", Secret: true}},
			Check: func(ctx context.Context, url string, cred map[string]string) error {
				in, err := domain.NewInstance("check", string(app), "", url, cred["apiKey"])
				if err != nil {
					return err
				}
				if err := provider.Check(ctx, in); err != nil {
					return kernel.SetupFailure(app.Title(), err)
				}
				return nil
			},
		})
	}
	return out
}

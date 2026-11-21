package adguard

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/adguard/domain"
	"waylight/internal/modules/adguard/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares AdGuard Home and Pi-hole as two services
// (#610). Pi-hole needs no user name; the check reads the statistics like
// SaveSettings.
func connectionSpecs(providers map[domain.Provider]usecase.DNSFilterProvider) []modules.ConnectionSpec {
	fields := map[domain.Provider][]modules.CredentialField{
		domain.ProviderAdGuard: {{Name: "username"}, {Name: "password", Secret: true}},
		domain.ProviderPiHole:  {{Name: "username", Optional: true}, {Name: "password", Secret: true}},
	}
	out := make([]modules.ConnectionSpec, 0, len(fields))
	for _, p := range []domain.Provider{domain.ProviderAdGuard, domain.ProviderPiHole} {
		provider := providers[p]
		out = append(out, modules.ConnectionSpec{
			Service: string(p), Title: p.Title(), Fields: fields[p],
			Check: func(ctx context.Context, url string, cred map[string]string) error {
				s, err := domain.NewSettings(p, url, cred["username"], cred["password"])
				if err != nil {
					return err
				}
				if _, err := provider.Stats(ctx, s); err != nil {
					return kernel.SetupFailure(p.Title(), err)
				}
				return nil
			},
		})
	}
	return out
}

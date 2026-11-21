package media

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/media/domain"
	"waylight/internal/modules/media/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares one service per media server (#610), so a
// widget can pick a Jellyfin and a Plex connection side by side. The check
// reads the overview like SaveSettings.
func connectionSpecs(providers map[domain.Provider]usecase.MediaServerProvider) []modules.ConnectionSpec {
	out := make([]modules.ConnectionSpec, 0, len(providers))
	for _, p := range []domain.Provider{domain.ProviderJellyfin, domain.ProviderPlex} {
		provider := providers[p]
		out = append(out, modules.ConnectionSpec{
			Service: string(p), Title: p.Title(),
			Fields: []modules.CredentialField{{Name: "token", Secret: true}},
			Check: func(ctx context.Context, url string, cred map[string]string) error {
				s, err := domain.NewSettings(string(p), url, cred["token"])
				if err != nil {
					return err
				}
				if _, err := provider.Overview(ctx, s); err != nil {
					return kernel.SetupFailure(p.Title(), err)
				}
				return nil
			},
		})
	}
	return out
}

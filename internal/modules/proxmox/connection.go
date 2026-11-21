package proxmox

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/proxmox/domain"
	"waylight/internal/modules/proxmox/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the Proxmox VE connection (#610). The token ID
// and the pinned certificate fingerprint are not secret; the check reads
// the cluster like SaveSettings.
func connectionSpecs(client usecase.Client) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "proxmox", Title: "Proxmox VE",
		Fields: []modules.CredentialField{{Name: "tokenId"}, {Name: "secret", Secret: true}, {Name: "fingerprint", Optional: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["tokenId"], cred["secret"], cred["fingerprint"])
			if err != nil {
				return err
			}
			if _, err := client.Cluster(ctx, s); err != nil {
				return kernel.SetupFailure("Proxmox VE", err)
			}
			return nil
		},
	}}
}

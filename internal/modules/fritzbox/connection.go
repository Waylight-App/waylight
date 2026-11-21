package fritzbox

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/fritzbox/domain"
	"waylight/internal/modules/fritzbox/usecase"
	"waylight/internal/shared/kernel"
)

// connectionSpecs declares the FRITZ!Box connection (#610). The login is
// optional: the internet connection is readable without it, the call list
// needs it. The check makes the same calls as SaveSettings.
func connectionSpecs(box usecase.Box) []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{
		Service: "fritzbox", Title: "FRITZ!Box",
		Fields: []modules.CredentialField{{Name: "username", Optional: true}, {Name: "password", Secret: true, Optional: true}},
		Check: func(ctx context.Context, url string, cred map[string]string) error {
			s, err := domain.NewSettings(url, cred["username"], cred["password"])
			if err != nil {
				return err
			}
			if _, _, err := box.WAN(ctx, s); err != nil {
				return kernel.SetupFailure("FRITZ!Box", err)
			}
			if s.HasLogin() {
				if _, err := box.Calls(ctx, s, 1); err != nil {
					return kernel.SetupFailure("FRITZ!Box", err)
				}
			}
			return nil
		},
	}}
}

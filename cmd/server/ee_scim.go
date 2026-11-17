//go:build ee

package main

import (
	"os"

	"waylight/ee/scim"
	"waylight/internal/modules"
)

// SCIM provisioning from a directory (#396), part of the ee/ features. Without
// SCIM_TOKEN every request is refused; without a licence key that includes
// directory-sync the routes answer 403.
func init() {
	registerModule("scim", func(d moduleDeps) modules.Module {
		return scim.New(d.db, d.auth, licenceModule(d).Licensing(), os.Getenv("SCIM_TOKEN"))
	})
}

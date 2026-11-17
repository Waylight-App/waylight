//go:build ee

package main

import (
	"waylight/ee/policies"
	"waylight/internal/modules"
)

// Admin policies (#304), part of the ee/ features. Without a licence key that
// includes policies no policy applies.
func init() {
	registerModule("policies", func(d moduleDeps) modules.Module {
		return policies.New(d.db, d.auth, d.registry, licenceModule(d).Licensing())
	})
}

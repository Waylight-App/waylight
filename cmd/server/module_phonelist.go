package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/phonelist"
)

// The phone list module registers itself; no shared file lists it (#193).
func init() {
	registerModule("phonelist", func(d moduleDeps) modules.Module {
		return phonelist.New(d.db, d.out.secrets, d.out.integration, d.clock, d.ids)
	})
}

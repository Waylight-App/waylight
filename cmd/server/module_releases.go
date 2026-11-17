package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/releases"
)

// The releases module registers itself; no shared file lists it (#193).
func init() {
	registerModule("releases", func(d moduleDeps) modules.Module { return releases.New(d.db, d.out.client, d.clock) })
}

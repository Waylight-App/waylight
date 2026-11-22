package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/meters"
)

// The meters module registers itself; no shared file lists it (#193).
func init() {
	registerModule("meters", func(d moduleDeps) modules.Module {
		return meters.New(d.db, d.clock, d.ids, time.Local)
	})
}

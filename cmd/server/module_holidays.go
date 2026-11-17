package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/holidays"
)

// The holidays module registers itself; no shared file lists it (#193).
func init() {
	registerModule("holidays", func(d moduleDeps) modules.Module {
		// The integration client re-checks the owner-configured ICS address on
		// every connection (#191).
		return holidays.New(d.db, d.out.integration, d.clock, time.Local)
	})
}

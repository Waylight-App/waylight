package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/birthdays"
)

// The birthdays module registers itself; no shared file lists it (#193).
func init() {
	registerModule("birthdays", func(d moduleDeps) modules.Module {
		return birthdays.New(d.db, d.out.secrets, d.out.integration, d.alerts, d.ids, d.clock, time.Local)
	})
}

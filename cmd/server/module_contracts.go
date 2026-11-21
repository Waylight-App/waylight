package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/contracts"
)

// The contracts module registers itself; no shared file lists it (#193).
// Reminders go to the notification channels through the alert relay.
func init() {
	registerModule("contracts", func(d moduleDeps) modules.Module {
		return contracts.New(d.db, d.clock, d.ids, d.alerts, time.Local)
	})
}

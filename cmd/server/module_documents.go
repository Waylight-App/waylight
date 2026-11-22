package main

import (
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/documents"
)

// The documents module registers itself; no shared file lists it (#193).
func init() {
	registerModule("documents", func(d moduleDeps) modules.Module {
		return documents.New(d.db, d.clock, d.ids, time.Local)
	})
}

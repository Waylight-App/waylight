package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/drop"
)

// The drop zone registers itself; no shared file lists it (#193).
func init() {
	registerModule("drop", func(d moduleDeps) modules.Module { return drop.New(d.db, d.clock, d.ids) })
}

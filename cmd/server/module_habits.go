package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/habits"
)

// The Habits module registers itself; no shared file lists it (#193).
func init() {
	registerModule("habits", func(d moduleDeps) modules.Module { return habits.New(d.db, d.clock, d.ids) })
}

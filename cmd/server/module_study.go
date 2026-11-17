package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/study"
)

// The study module registers itself; no shared file lists it (#193).
func init() {
	registerModule("study", func(d moduleDeps) modules.Module { return study.New(d.db, d.ids) })
}

package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/phonelist"
)

// The Phone list module registers itself; no shared file lists it (#193).
func init() {
	registerModule("phonelist", func(d moduleDeps) modules.Module { return phonelist.New(d.db, d.clock, d.ids) })
}

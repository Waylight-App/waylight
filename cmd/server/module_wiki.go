package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/wiki"
)

// The Wiki search module registers itself; no shared file lists it (#193).
func init() {
	registerModule("wiki", func(d moduleDeps) modules.Module {
		return wiki.New(d.db, d.out.secrets, d.out.integration, d.clock, d.ids)
	})
}

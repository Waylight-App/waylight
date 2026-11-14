package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/sports"
)

// The sports module registers itself; no shared file lists it (#193).
func init() {
	registerModule("sports", func(d moduleDeps) modules.Module {
		return sports.New(d.out.client, d.cfg.SportsAPIURL, d.clock)
	})
}

package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/electricity"
)

// The electricity price module registers itself; no shared file lists it (#193).
func init() {
	registerModule("electricity", func(d moduleDeps) modules.Module {
		return electricity.New(d.out.client, d.cfg.ElectricityAPIURL, d.clock)
	})
}

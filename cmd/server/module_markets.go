package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/markets"
)

// The markets module registers itself; no shared file lists it (#193).
func init() {
	registerModule("markets", func(d moduleDeps) modules.Module {
		return markets.New(d.out.client, d.cfg.MarketsECBURL, d.cfg.MarketsStooqURL, d.clock)
	})
}

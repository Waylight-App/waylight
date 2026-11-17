package main

import (
	"waylight/internal/modules"
	"waylight/internal/modules/markets"
)

// The markets module registers itself; no shared file lists it (#193).
func init() {
	registerModule("markets", func(d moduleDeps) modules.Module {
		return markets.New(d.out.client, d.cfg.MarketsECBURL, d.cfg.MarketsStooqURL, d.clock, d.cfg.CurrencyRates)
	})
}

//go:build ee

package main

import (
	"crypto/ed25519"
	"os"
	"sync"

	"trisse-home/ee/licence"
	"trisse-home/internal/modules"
)

// The licence module of the ee/ features (#307). This file is only built with
// the "ee" tag, so a build without it contains no code from ee/ (ADR 0017).

var (
	licenceMu   sync.Mutex
	licenceMods = map[*modules.Registry]*licence.Module{}
	// licencePublicKey replaces the compiled-in key in tests; nil in production.
	licencePublicKey ed25519.PublicKey
)

// licenceModule returns the licence module of the server being built, creating
// it on first use. Other ee/ modules are built in alphabetical order, some
// before "licence", so they get the port from here instead of the registry:
//
//	registerModule("branding", func(d moduleDeps) modules.Module {
//		return branding.New(d.db, licenceModule(d).Licensing())
//	})
//
// It is kept per registry because tests build several servers in one process.
func licenceModule(d moduleDeps) *licence.Module {
	licenceMu.Lock()
	defer licenceMu.Unlock()
	m, ok := licenceMods[d.registry]
	if !ok {
		m = licence.New(d.db, d.clock, d.auth, licence.Config{EnvKey: os.Getenv("LICENCE_KEY"), PublicKey: licencePublicKey})
		licenceMods[d.registry] = m
	}
	return m
}

func init() {
	registerModule("licence", func(d moduleDeps) modules.Module { return licenceModule(d) })
}

//go:build ee

package main

import (
	"log/slog"

	"waylight/ee/audit"
	"waylight/internal/modules"
)

// The audit log (#303), part of the ee/ features. It receives every change to
// instance data and sharing through changeSink; without a licence key that
// includes audit-log it records nothing and its routes answer 403.
func init() {
	registerModule("audit", func(d moduleDeps) modules.Module {
		m := audit.New(d.db, d.clock, d.ids, licenceModule(d).Licensing(), slog.Default())
		changeSink = m
		return m
	})
}

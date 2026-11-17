//go:build ee

package main

import (
	"context"
	"log/slog"
	"sync"

	"waylight/ee/absences"
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

var (
	absencesMu   sync.Mutex
	absencesMods = map[*modules.Registry]*absences.Module{}
)

// absencesModule returns the absences module of a server; presence (#346)
// reads planned absences from the same instance. It is kept per registry
// because tests build several servers in one process.
func absencesModule(d moduleDeps) *absences.Module {
	absencesMu.Lock()
	defer absencesMu.Unlock()
	m, ok := absencesMods[d.registry]
	if !ok {
		m = newAbsences(d)
		absencesMods[d.registry] = m
	}
	return m
}

// Absences (#299), part of the ee/ features.
func init() {
	registerModule("absences", func(d moduleDeps) modules.Module { return absencesModule(d) })
}

func newAbsences(d moduleDeps) *absences.Module {
	m := absences.New(absences.Deps{
		DB: d.db, Licensing: licenceModule(d).Licensing(), Names: accountNames{d.auth.Provisioning()}, IDs: d.ids, Clock: d.clock,
		Audit: func(ctx context.Context, action, id string) {
			// Only the action and the ID: kind and note can be health data.
			if changeSink != nil {
				p := kernel.PrincipalFrom(ctx)
				changeSink.RecordChange(ctx, modules.Change{AccountID: p.AccountID, Role: p.Role, Module: "absences",
					Route: action, ObjectID: id})
			}
		},
	})
	d.auth.OnAccountDeleted(func(ctx context.Context, accountID string) {
		if err := m.ForgetAccount(ctx, accountID); err != nil {
			slog.Default().Warn("absences: remove a deleted account's absences", "err", err)
		}
	})
	return m
}

//go:build ee

package main

import (
	"context"
	"log/slog"

	"trisse-home/ee/absences"
	"trisse-home/internal/modules"
	"trisse-home/internal/shared/kernel"
)

// Absences (#299), part of the ee/ features.
func init() {
	registerModule("absences", func(d moduleDeps) modules.Module {
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
	})
}

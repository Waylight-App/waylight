//go:build ee

package main

import (
	"context"

	"trisse-home/ee/branding"
	"trisse-home/internal/modules"
	"trisse-home/internal/shared/kernel"
)

// Branding (#305), part of the ee/ features: the module becomes the core's
// branding source; without a licence key that includes branding the
// product's branding is shown.
func init() {
	registerModule("branding", func(d moduleDeps) modules.Module {
		m := branding.New(branding.Deps{DB: d.db, Licensing: licenceModule(d).Licensing(), Clock: d.clock,
			Audit: func(ctx context.Context, action, id string) {
				if changeSink != nil {
					p := kernel.PrincipalFrom(ctx)
					changeSink.RecordChange(ctx, modules.Change{AccountID: p.AccountID, Role: p.Role, Module: "branding", Route: action, ObjectID: id})
				}
			}})
		brandingSource = m.Source()
		return m
	})
}

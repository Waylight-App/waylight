// Package teamtabs lets admins share their own dashboard tabs with a group
// or everyone (#295, part of the ee/ features). Members of the group see the
// tab next to their own, read-only. The dashboard shows shared tabs only
// while this feature is licensed.
//
// Ring: module wiring.
package teamtabs

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/ee/licence"
	"waylight/ee/teamtabs/adapter/controller"
	"waylight/ee/teamtabs/adapter/gateway"
	"waylight/ee/teamtabs/usecase"
	"waylight/internal/modules"
)

// Module wires the rings.
type Module struct {
	modules.Base
	svc       *usecase.Service
	handlers  *controller.Handlers
	licensing licence.Licensing
}

// New builds the module.
func New(db *sql.DB, sharing usecase.Sharing, l licence.Licensing) *Module {
	svc := usecase.New(sharing, gateway.NewRepository(db))
	return &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: l}
}

// DefaultTab is the team tab that opens by default for the principal's
// groups ("" for none); the dashboard asks it while team tabs are on.
func (m *Module) DefaultTab(ctx context.Context) string { return m.svc.DefaultFor(ctx) }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// Access implements modules.Gated: with an expired key team tabs stay
// visible, but nothing can be shared or unshared.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureTeamTabs)
}

// SharedTabsOn reports whether members see tabs shared with them.
func (m *Module) SharedTabsOn(ctx context.Context) bool { return m.Access(ctx) != modules.AccessNone }

// Name implements modules.Module.
func (m *Module) Name() string { return "teamtabs" }

// DataKind implements modules.Module: the dashboard holds the tabs; this
// module only changes who sees an admin's tab.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/team-tabs", m.handlers.List)
	r.Protected("GET /api/team-tabs/defaults", m.handlers.Defaults)
	r.Protected("PUT /api/team-tabs/defaults/{groupId}", m.handlers.SetDefault)
	r.Protected("PUT /api/team-tabs/{id}", m.handlers.Share)
}

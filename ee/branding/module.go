// Package branding lets admins give the instance their organisation's name,
// logos, favicon and colours (#305, part of the ee/ features). The core
// shows them through its branding Source port; this module only stores and
// supplies them while the licence covers branding.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic.
package branding

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/ee/branding/adapter/controller"
	"trisse-home/ee/branding/adapter/gateway"
	"trisse-home/ee/branding/usecase"
	"trisse-home/ee/licence"
	"trisse-home/internal/modules"
	"trisse-home/internal/shared/kernel"
)

// Module wires the rings.
type Module struct {
	modules.Base
	svc       *usecase.Service
	handlers  *controller.Handlers
	licensing licence.Licensing
}

// Deps are what the module needs from the server.
type Deps struct {
	DB        *sql.DB
	Licensing licence.Licensing
	Clock     kernel.Clock
	Audit     func(ctx context.Context, action, id string)
}

// New builds the module.
func New(d Deps) *Module {
	svc := usecase.New(usecase.Deps{Store: gateway.NewStore(d.DB), Licensing: d.Licensing, Clock: d.Clock, Audit: d.Audit})
	return &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: d.Licensing}
}

// Source is the organisation's branding for the core (branding.Source).
func (m *Module) Source() *usecase.Service { return m.svc }

// Access implements modules.Gated: with an expired key admins still see
// the stored branding, but cannot change it.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureBranding)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "branding" }

// DataKind implements modules.Module.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module; admins only (use cases).
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/branding/settings", h.Get)
	r.Protected("PUT /api/branding/settings", h.Save)
	r.Protected("PUT /api/branding/assets/{kind}", h.Upload)
	r.Protected("DELETE /api/branding/assets/{kind}", h.Delete)
}

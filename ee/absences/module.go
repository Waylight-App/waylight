// Package absences shows who is away (#299, part of the ee/ features):
// everyone enters their own holidays, sick days and remote days, admins
// enter them for anyone, and a private absence shows others only "away".
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic.
package absences

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/ee/absences/adapter/controller"
	"trisse-home/ee/absences/adapter/gateway"
	"trisse-home/ee/absences/usecase"
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
	Names     usecase.Names
	IDs       kernel.IDGenerator
	Clock     kernel.Clock
	Audit     func(ctx context.Context, action, id string)
}

// New builds the module.
func New(d Deps) *Module {
	svc := usecase.New(usecase.Deps{Store: gateway.NewStore(d.DB), Names: d.Names, IDs: d.IDs, Clock: d.Clock, Audit: d.Audit})
	return &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: d.Licensing}
}

// ForgetAccount removes a deleted account's absences.
func (m *Module) ForgetAccount(ctx context.Context, accountID string) error {
	return m.svc.ForgetAccount(ctx, accountID)
}

// Access implements modules.Gated: readable with an expired key, no new
// entries.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureAbsences)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "absences" }

// DataKind implements modules.Module. Members write their own absences;
// the use cases decide whose, so the central guard lets members' writes
// through like personal settings.
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/absences", h.List)
	r.Protected("POST /api/absences", h.Create)
	r.Protected("PUT /api/absences/{id}", h.Update)
	r.Protected("DELETE /api/absences/{id}", h.Delete)
}

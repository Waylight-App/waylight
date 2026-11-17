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

	"waylight/ee/absences/adapter/controller"
	"waylight/ee/absences/adapter/gateway"
	"waylight/ee/absences/usecase"
	"waylight/ee/licence"
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
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

// AwayOn is a read model for the presence module (#346), wired in
// cmd/server: who is away on a day, without kind or note.
func (m *Module) AwayOn(ctx context.Context, day string) ([]usecase.Away, error) {
	return m.svc.AwayOn(ctx, day)
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

// WidgetKinds implements modules.Module: the widget of the ee/ plugin
// (ADR 0021, #299); it disappears from the picker with the licence.
func (m *Module) WidgetKinds() []string { return []string{"absences"} }

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

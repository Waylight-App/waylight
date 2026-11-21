package places

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/places/adapter/controller"
	"waylight/internal/modules/places/adapter/gateway"
	"waylight/internal/modules/places/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the Places rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// Widgets lists the widgets on the account's own tabs (the dashboard), so a
// place can say which widgets use it.
type Widgets func(ctx context.Context) ([]usecase.Widget, error)

// New builds the places module (#591).
func New(db *sql.DB, widgets Widgets, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewPlaces(usecase.Deps{Repo: gateway.NewRepository(db), Widgets: widgets, Clock: clock, IDs: ids})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "places" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind says places belong to the account that saved them and are
// shared like other personal content (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/places", m.handlers.List)
	r.Protected("POST /api/places", m.handlers.Add)
	r.Protected("PUT /api/places/{id}", m.handlers.Update)
	r.Protected("DELETE /api/places/{id}", m.handlers.Delete)
}

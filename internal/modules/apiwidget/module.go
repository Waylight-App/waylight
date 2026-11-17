package apiwidget

import (
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/apiwidget/adapter/controller"
	"waylight/internal/modules/apiwidget/adapter/gateway"
	"waylight/internal/modules/apiwidget/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the JSON widget rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// Deps are the dependencies of the module. Fetcher holds the two guarded
// HTTP clients (private addresses only, and any non-link-local address).
type Deps struct {
	DB      *sql.DB
	Cipher  gateway.Cipher
	Fetcher gateway.Fetcher
	Clock   kernel.Clock
	IDs     kernel.IDGenerator
}

// New builds the module (#48). Nothing is requested until an endpoint is
// configured and a widget shows it.
func New(d Deps) *Module {
	uc := usecase.New(gateway.NewRepository(d.DB, d.Cipher), d.Fetcher, gateway.JSONPath{}, d.Clock, d.IDs)
	return &Module{
		Transferable: modules.Transferable{DB: d.DB, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(uc),
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "apiwidget" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"apiwidget"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/apiwidget/endpoints", m.handlers.List)
	r.Protected("POST /api/apiwidget/endpoints", m.handlers.Create)
	r.Protected("PUT /api/apiwidget/endpoints/{id}", m.handlers.Update)
	r.Protected("DELETE /api/apiwidget/endpoints/{id}", m.handlers.Delete)
	r.Protected("GET /api/apiwidget/endpoints/{id}/values", m.handlers.Values)
	r.Protected("POST /api/apiwidget/preview", m.handlers.Preview)
}

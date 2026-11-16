package phonelist

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/phonelist/adapter/controller"
	"trisse-home/internal/modules/phonelist/adapter/gateway"
	"trisse-home/internal/modules/phonelist/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the phone list rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the phone list module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewPhonelist(usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "phonelist" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind implements modules.Module (ADR 0018): admins maintain one list
// for everyone; members and guests read it.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"phonelist"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/phonelist/entries", m.handlers.List)
	r.Protected("POST /api/phonelist/entries", m.handlers.Add)
	r.Protected("PUT /api/phonelist/entries/{id}", m.handlers.Update)
	r.Protected("DELETE /api/phonelist/entries/{id}", m.handlers.Delete)
}

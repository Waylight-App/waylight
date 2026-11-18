package habits

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/habits/adapter/controller"
	"waylight/internal/modules/habits/adapter/gateway"
	"waylight/internal/modules/habits/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the habit tracker rings and implements modules.Module (#106).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the habit tracker module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids}
	uc := usecase.NewHabits(deps)
	return &Module{seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) }, Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "habits" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"habits"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/habits", m.handlers.List)
	r.Protected("POST /api/habits", m.handlers.Create)
	r.Protected("PUT /api/habits/{id}", m.handlers.Update)
	r.Protected("DELETE /api/habits/{id}", m.handlers.Delete)
	r.Protected("POST /api/habits/{id}/toggle", m.handlers.Toggle)
}

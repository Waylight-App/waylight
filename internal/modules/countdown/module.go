package countdown

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/countdown/adapter/controller"
	"waylight/internal/modules/countdown/adapter/gateway"
	"waylight/internal/modules/countdown/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the countdown rings and implements modules.Module (#54).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the countdown module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids}
	uc := usecase.NewCountdowns(deps)
	return &Module{seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) }, Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "countdown" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"countdown"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/countdowns", m.handlers.List)
	r.Protected("POST /api/countdowns", m.handlers.Create)
	r.Protected("PUT /api/countdowns/{id}", m.handlers.Update)
	r.Protected("DELETE /api/countdowns/{id}", m.handlers.Delete)
}

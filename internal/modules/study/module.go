package study

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/study/adapter/controller"
	"waylight/internal/modules/study/adapter/gateway"
	"waylight/internal/modules/study/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the study rings and implements modules.Module.
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// New builds the module (#57); it makes no outbound requests.
func New(db *sql.DB, ids kernel.IDGenerator) *Module {
	repo := gateway.NewRepository(db)
	uc := usecase.New(repo, ids)
	return &Module{seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, repo, ids) }, Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "study" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"study"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/study/modules", m.handlers.List)
	r.Protected("POST /api/study/modules", m.handlers.Create)
	r.Protected("PUT /api/study/modules/{id}", m.handlers.Update)
	r.Protected("DELETE /api/study/modules/{id}", m.handlers.Delete)
	r.Protected("GET /api/study/progress", m.handlers.Progress)
	r.Protected("GET /api/study/settings", m.handlers.Settings)
	r.Protected("PUT /api/study/settings", m.handlers.SaveSettings)
}

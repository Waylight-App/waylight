package focus

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"waylight/internal/modules/focus/domain"

	"waylight/internal/modules"
	"waylight/internal/modules/focus/adapter/controller"
	"waylight/internal/modules/focus/adapter/gateway"
	"waylight/internal/modules/focus/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the focus timer rings and implements modules.Module (#58).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	uc       usecase.Focus
	clock    kernel.Clock
}

// New builds the focus timer module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids}
	uc := usecase.NewFocus(deps)
	return &Module{seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) }, Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc), uc: uc, clock: clock}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "focus" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"focus"} }

// Jobs implements modules.Module: old sessions are pruned once a day.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "focus.prune", Interval: 24 * time.Hour, Run: m.uc.Prune}}
}

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/focus/stats", m.handlers.Stats)
	r.Protected("POST /api/focus/sessions", m.handlers.Record)
}

// CompactState implements modules.CompactProvider (#597): today's focus
// time; it is neither good nor bad, so the level stays none.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	days, err := m.uc.WeeklyStats(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	var today domain.Day
	if len(days) > 0 {
		today = days[len(days)-1]
	}
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: m.clock.Now(),
		Title: kernel.NewTextf("{minutes} min focus today", "minutes", today.Minutes)}
	if today.Sessions > 0 {
		state.Detail = kernel.NewTextf("{count} sessions", "count", today.Sessions)
	}
	return state, nil
}

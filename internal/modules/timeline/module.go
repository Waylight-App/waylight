package timeline

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/timeline/adapter/controller"
	"waylight/internal/modules/timeline/adapter/gateway"
	"waylight/internal/modules/timeline/domain"
	"waylight/internal/modules/timeline/usecase"
	"waylight/internal/shared/kernel"
)

// pruneInterval: old events are removed once a day.
const pruneInterval = 24 * time.Hour

// Module wires the timeline rings, implements modules.Module and receives
// observations and alerts like the notifications module.
type Module struct {
	modules.Base
	uc       usecase.Timeline
	handlers *controller.Handlers
	seedDemo func(ctx context.Context) error
}

// New builds the timeline module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	repo := gateway.NewRepository(db)
	uc := usecase.NewTimeline(repo, clock, ids)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc),
		seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, repo, clock, ids) }}
}

// Observe implements kernel.ObservationSink.
func (m *Module) Observe(ctx context.Context, o kernel.Observation) { m.uc.Observe(ctx, o) }

// Alert implements kernel.AlertSink.
func (m *Module) Alert(ctx context.Context, a kernel.Alert) { m.uc.Alert(ctx, a) }

// SeedDemo implements modules.DemoSeeder: a plausible day of events.
func (m *Module) SeedDemo(ctx context.Context) error { return m.seedDemo(ctx) }

// Name implements modules.Module.
func (m *Module) Name() string { return "timeline" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"timeline"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/timeline", m.handlers.List)
	r.Protected("GET /api/timeline/settings", m.handlers.Settings)
	r.Protected("PUT /api/timeline/settings", m.handlers.SaveSettings)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "timeline.prune", Interval: pruneInterval, Run: m.uc.Prune}}
}

// Recent returns up to limit events, newest first (read access for the
// briefing, ADR 0016).
func (m *Module) Recent(ctx context.Context, limit int) ([]domain.Event, error) {
	return m.uc.List(ctx, usecase.Filter{Limit: limit})
}

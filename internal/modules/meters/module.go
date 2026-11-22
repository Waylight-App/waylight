package meters

import (
	"context"
	"database/sql"
	"io/fs"
	"slices"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/meters/adapter/controller"
	"waylight/internal/modules/meters/adapter/gateway"
	"waylight/internal/modules/meters/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the meter rings and implements modules.Module (#689).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	handlers *controller.Handlers
	uc       usecase.Meters
	clock    kernel.Clock
	loc      *time.Location
}

// New builds the meters module; loc decides what "today" is.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, loc *time.Location) *Module {
	if loc == nil {
		loc = time.Local
	}
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids, Location: loc}
	uc := usecase.NewMeters(deps)
	return &Module{
		seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) },
		handlers: controller.NewHandlers(uc),
		uc:       uc,
		clock:    clock,
		loc:      loc,
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "meters" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"meters"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/meters", m.handlers.List)
	r.Protected("POST /api/meters", m.handlers.Create)
	r.Protected("GET /api/meters/{id}", m.handlers.Detail)
	r.Protected("PUT /api/meters/{id}", m.handlers.Update)
	r.Protected("DELETE /api/meters/{id}", m.handlers.Delete)
	r.Protected("POST /api/meters/{id}/readings", m.handlers.AddReading)
	r.Protected("DELETE /api/meters/{id}/readings/{day}", m.handlers.DeleteReading)
}

// CompactState implements modules.CompactProvider (#597): the days since the
// last reading of the meter that was read longest ago.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	list, err := m.uc.List(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: m.clock.Now()}
	if len(list) == 0 {
		state.Title = kernel.NewText("No meters yet")
		state.Empty = true
		return state, nil
	}
	read := slices.DeleteFunc(slices.Clone(list), func(v usecase.MeterView) bool { return v.Last == nil })
	if len(read) == 0 {
		state.Title = kernel.NewText("No readings yet")
		state.Empty = true
		return state, nil
	}
	oldest := slices.MinFunc(read, func(a, b usecase.MeterView) int { return a.Last.Day.Compare(b.Last.Day) })
	n := m.clock.Now().In(m.loc)
	days := int(time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC).Sub(oldest.Last.Day).Hours() / 24)
	switch {
	case days <= 0:
		state.Title = kernel.NewTextf("{name}: read today", "name", oldest.Name)
	case days == 1:
		state.Title = kernel.NewTextf("{name}: read 1 day ago", "name", oldest.Name)
	default:
		state.Title = kernel.NewTextf("{name}: read {days} days ago", "name", oldest.Name, "days", days)
	}
	return state, nil
}

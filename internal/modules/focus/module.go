package focus

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
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
	repo := gateway.NewRepository(db)
	deps := usecase.Deps{Repo: repo, Timers: repo, Clock: clock, IDs: ids}
	uc := usecase.NewFocus(deps)
	return &Module{seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) }, Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc, clock.Now), uc: uc, clock: clock}
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
	r.Protected("GET /api/focus/timer", m.handlers.Timer)
	r.Protected("PUT /api/focus/timer/config", m.handlers.ConfigureTimer)
	r.Protected("POST /api/focus/timer/{action}", m.handlers.ControlTimer)
}

// CompactState implements modules.CompactProvider (#597): the running
// timer with Pause, or today's focus time with Start (#631). Focus time is
// neither good nor bad, so the level stays none.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	days, err := m.uc.WeeklyStats(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	var today domain.Day
	if len(days) > 0 {
		today = days[len(days)-1]
	}
	timer, err := m.uc.Timer(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	now := m.clock.Now()
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: now, Title: timerTitle(timer, now, today)}
	if today.Sessions > 0 {
		state.Detail = kernel.NewTextf("{count} sessions", "count", today.Sessions)
	}
	if p := kernel.PrincipalFrom(ctx); !p.TabScoped() {
		state.Action = timerAction(timer)
	}
	return state, nil
}

// timerTitle is "Focus · 12:40" while a phase runs or is paused part-way,
// else today's focus time.
func timerTitle(t domain.Timer, now time.Time, today domain.Day) kernel.Text {
	clock := domain.FormatClock(t.Remaining(now))
	switch {
	case t.Untouched():
		return kernel.NewTextf("{minutes} min focus today", "minutes", today.Minutes)
	case !t.Running():
		return kernel.NewTextf("Paused · {time}", "time", clock)
	case t.Phase == domain.PhaseShort:
		return kernel.NewTextf("Short break · {time}", "time", clock)
	case t.Phase == domain.PhaseLong:
		return kernel.NewTextf("Long break · {time}", "time", clock)
	default:
		return kernel.NewTextf("Focus · {time}", "time", clock)
	}
}

// timerAction is Pause while the timer runs and Start otherwise.
func timerAction(t domain.Timer) *kernel.CompactAction {
	if t.Running() {
		return &kernel.CompactAction{ID: "focus.pause", Label: kernel.NewText("Pause"), Method: http.MethodPost, Path: "/api/focus/timer/pause"}
	}
	return &kernel.CompactAction{ID: "focus.start", Label: kernel.NewText("Start"), Method: http.MethodPost, Path: "/api/focus/timer/start"}
}

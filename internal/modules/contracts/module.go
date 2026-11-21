package contracts

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/contracts/adapter/controller"
	"waylight/internal/modules/contracts/adapter/gateway"
	"waylight/internal/modules/contracts/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the contract rings and implements modules.Module (#687).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	uc       usecase.Contracts
	clock    kernel.Clock
}

// New builds the contract module; alerts receives the reminders (nil =
// none) and loc decides what "today" is.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, alerts kernel.AlertSink, loc *time.Location) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids, Location: loc, Alerts: alerts}
	uc := usecase.NewContracts(deps)
	return &Module{
		seedDemo:     func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) },
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(uc),
		uc:           uc,
		clock:        clock,
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "contracts" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"contracts"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/contracts", m.handlers.List)
	r.Protected("POST /api/contracts", m.handlers.Create)
	r.Protected("PUT /api/contracts/{id}", m.handlers.Update)
	r.Protected("DELETE /api/contracts/{id}", m.handlers.Delete)
}

// Jobs implements modules.Module: the reminder check runs hourly; each
// deadline is reminded once, so the interval only decides how soon after
// midnight the reminder arrives.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "contracts.remind", Interval: time.Hour, Run: func(ctx context.Context) error {
		_, err := m.uc.Remind(ctx)
		return err
	}}}
}

// CompactState implements modules.CompactProvider (#597): the next
// cancellation deadline; attention inside its reminder period.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	list, err := m.uc.List(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: m.clock.Now()}
	soon := 0
	for _, c := range list {
		if c.Soon {
			soon++
		}
	}
	if len(list) == 0 || !list[0].Term.Open {
		state.Title = kernel.NewText("No deadlines ahead")
		state.Empty = true
		return state, nil
	}
	next := list[0]
	state.Title = kernel.NewTextf("{name}: cancel by {date}", "name", next.Name, "date", next.Term.Deadline.Format(time.DateOnly))
	if soon > 0 {
		state.Level = kernel.CompactAttention
		state.Count = kernel.IntPtr(soon)
	}
	return state, nil
}

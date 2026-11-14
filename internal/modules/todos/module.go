package todos

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/todos/adapter/controller"
	"trisse-home/internal/modules/todos/adapter/gateway"
	"trisse-home/internal/modules/todos/domain"
	"trisse-home/internal/modules/todos/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the to-do rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers  *controller.Handlers
	seed      usecase.SeedSampleData
	reminders usecase.SendDueReminders
	list      usecase.ListTodos
}

// reminderInterval is how often due reminders are checked (#68).
const reminderInterval = time.Minute

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "todos.reminders", Interval: reminderInterval, Run: func(ctx context.Context) error {
		_, err := m.reminders.Execute(ctx)
		return err
	}}}
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

// New builds the to-do module. Reminders are delivered to alerts (the
// notifications module, kind "reminder") in the server's time zone.
// External lists (#169) use client for CalDAV and cipher for their
// passwords.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, alerts kernel.AlertSink, cipher gateway.Cipher, client *http.Client) *Module {
	d := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids, Location: time.Local, Alerts: alerts,
		Sources: gateway.NewSourceRepository(db, cipher), Providers: gateway.NewProviders(client)}
	list := usecase.NewListTodos(d)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, seed: usecase.NewSeedSampleData(d),
		reminders: usecase.NewSendDueReminders(d), list: list, handlers: controller.NewHandlers(controller.UseCases{
			List: list, Create: usecase.NewCreateTodo(d), Update: usecase.NewUpdateTodo(d),
			Delete: usecase.NewDeleteTodo(d), ClearCompleted: usecase.NewClearCompleted(d), Sources: usecase.NewSources(d),
		})}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "todos" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"todos"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/todos", h.List)
	r.Protected("POST /api/todos", h.Create)
	r.Protected("POST /api/todos/clear-completed", h.ClearCompleted)
	r.Protected("PUT /api/todos/{id}", h.Update)
	r.Protected("DELETE /api/todos/{id}", h.Delete)
	r.Protected("GET /api/todos/settings/sources", h.Sources)
	r.Protected("POST /api/todos/settings/sources", h.CreateSource)
	r.Protected("POST /api/todos/settings/sources/discover", h.DiscoverSources)
	r.Protected("PUT /api/todos/settings/sources/{id}", h.UpdateSource)
	r.Protected("DELETE /api/todos/settings/sources/{id}", h.DeleteSource)
}

// OpenTodos returns the open to-dos (read access for the briefing, wired
// in the composition root – ADR 0016).
func (m *Module) OpenTodos(ctx context.Context) ([]usecase.TodoView, error) {
	return m.list.Execute(ctx, domain.LocalSource, false)
}

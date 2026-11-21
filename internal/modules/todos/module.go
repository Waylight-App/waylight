package todos

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"net/url"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/todos/adapter/controller"
	"waylight/internal/modules/todos/adapter/gateway"
	"waylight/internal/modules/todos/domain"
	"waylight/internal/modules/todos/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the to-do rings and implements modules.Module.
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	modules.Transferable
	handlers  *controller.Handlers
	seed      usecase.SeedSampleData
	reminders usecase.SendDueReminders
	list      usecase.ListTodos
	clock     kernel.Clock
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
// publicURL (PUBLIC_URL) enables signing in to Microsoft, Google and
// TickTick (#171); nil leaves those kinds unavailable.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, alerts kernel.AlertSink, cipher gateway.Cipher, client *http.Client, publicURL *url.URL) *Module {
	local := gateway.NewRepository(db)
	d := usecase.Deps{Repo: local, Visibility: local, Clock: clock, IDs: ids, Location: time.Local, Alerts: alerts,
		Sources: gateway.NewSourceRepository(db, cipher), OAuth: newOAuthPort(client, publicURL, clock.Now)}
	d.Providers = gateway.NewProviders(client, usecase.NewTokenSource(d))
	list := usecase.NewListTodos(d)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, seed: usecase.NewSeedSampleData(d), seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, d) },
		reminders: usecase.NewSendDueReminders(d), list: list, clock: clock, handlers: controller.NewHandlers(controller.UseCases{
			List: list, Create: usecase.NewCreateTodo(d), Update: usecase.NewUpdateTodo(d),
			Delete: usecase.NewDeleteTodo(d), ClearCompleted: usecase.NewClearCompleted(d), Sources: usecase.NewSources(d),
			Share: usecase.NewShareTodo(d),
		})}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "todos" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

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
	r.Protected("PUT /api/todos/{id}/visibility", h.Share)
	r.Protected("GET /api/todos/settings/sources", h.Sources)
	r.Protected("POST /api/todos/settings/sources", h.CreateSource)
	r.Protected("POST /api/todos/settings/sources/discover", h.DiscoverSources)
	r.Protected("PUT /api/todos/settings/sources/{id}", h.UpdateSource)
	r.Protected("DELETE /api/todos/settings/sources/{id}", h.DeleteSource)
	r.Protected("GET /api/todos/settings/oauth", h.OAuthInfo)
	r.Protected("POST /api/todos/settings/oauth/start", h.StartOAuth)
	// Public: the session cookie is SameSite=Strict and does not come along
	// on the provider's redirect. The unguessable state of a sign-in that a
	// signed-in user started binds the answer, as in the SSO callback.
	r.Public("GET "+CallbackPath, h.OAuthCallback)
	r.Protected("GET /api/todos/settings/oauth/state/{state}", h.OAuthStatus)
}

// OpenTodos returns the open to-dos (read access for the briefing, wired
// in the composition root – ADR 0016).
func (m *Module) OpenTodos(ctx context.Context) ([]usecase.TodoView, error) {
	return m.list.Execute(ctx, domain.LocalSource, false)
}

// CompactState implements modules.CompactProvider (#597): open to-dos, and
// a look is asked for while some are overdue.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	list, err := m.list.Execute(ctx, "", false)
	if err != nil {
		return kernel.CompactState{}, err
	}
	open, overdue := 0, 0
	for _, t := range list {
		if t.Done {
			continue
		}
		open++
		if t.Overdue {
			overdue++
		}
	}
	state := kernel.CompactState{Level: kernel.CompactOK, Count: kernel.IntPtr(open), UpdatedAt: m.clock.Now()}
	state.Title = kernel.NewTextf("{count} open", "count", open)
	if open == 0 {
		state.Title = kernel.NewText("Nothing to do")
	}
	if overdue > 0 {
		state.Level = kernel.CompactAttention
		state.Detail = kernel.NewTextf("{count} overdue", "count", overdue)
	}
	return state, nil
}

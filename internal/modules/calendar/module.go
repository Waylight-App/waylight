package calendar

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"net/url"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/calendar/adapter/controller"
	"trisse-home/internal/modules/calendar/adapter/gateway"
	"trisse-home/internal/modules/calendar/usecase"
	"trisse-home/internal/shared/kernel"
)

const refreshInterval = 15 * time.Minute

// Module wires the calendar rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Service
	handlers *controller.Handlers
	subs     *controller.SubscriptionHandlers
	accounts *controller.AccountHandlers
}

// New builds the calendar module. Floating event times are interpreted in loc.
// publicURL (PUBLIC_URL) enables signing in to Google and Microsoft (#282);
// nil leaves that unavailable.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator, loc *time.Location, publicURL *url.URL,
	sharing *Sharing) *Module {
	caldav := gateway.NewCalDAV(client, loc)
	subs := gateway.NewSubscriptionRepository(db, cipher)
	if sharing == nil {
		sharing = NewSharing()
	}
	sharing.repo = subs
	uc := usecase.NewCalendar(usecase.Deps{
		Shared: subs, SharingOn: sharing.on,
		Repo: gateway.NewSettingsRepository(db, cipher), Source: caldav, Writer: caldav,
		Subscriptions: subs, Feeds: gateway.NewFeed(client, loc, clock.Now),
		IDs: ids, Clock: clock,
		Accounts: gateway.NewAccountRepository(db, cipher), API: gateway.NewAccountAPI(client), OAuth: newOAuthPort(client, publicURL, clock.Now),
	})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), subs: controller.NewSubscriptionHandlers(uc), accounts: controller.NewAccountHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "calendar" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"calendar"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/calendar/settings", m.handlers.Settings)
	r.Protected("PUT /api/calendar/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/calendar/events", m.handlers.Events)
	r.Protected("POST /api/calendar/events", m.handlers.CreateEvent)
	r.Protected("GET /api/calendar/subscriptions", m.subs.List)
	r.Protected("POST /api/calendar/subscriptions", m.subs.Add)
	r.Protected("PUT /api/calendar/subscriptions/{id}", m.subs.Update)
	r.Protected("DELETE /api/calendar/subscriptions/{id}", m.subs.Delete)
	r.Protected("GET /api/calendar/writable", m.accounts.Writable)
	r.Protected("GET /api/calendar/accounts", m.accounts.List)
	r.Protected("POST /api/calendar/accounts/oauth/start", m.accounts.Start)
	r.Protected("GET /api/calendar/accounts/oauth/state/{state}", m.accounts.Status)
	r.Public("GET "+CallbackPath, m.accounts.Callback)
	r.Protected("PUT /api/calendar/accounts/{id}", m.accounts.SetCalendars)
	r.Protected("DELETE /api/calendar/accounts/{id}", m.accounts.Disconnect)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "calendar.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}

// Upcoming returns the events of the next days (read access for the
// briefing, ADR 0016).
func (m *Module) Upcoming(ctx context.Context, days int) (usecase.AgendaResponse, error) {
	return m.uc.Upcoming(ctx, days)
}

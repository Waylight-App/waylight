package birthdays

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/birthdays/adapter/controller"
	"waylight/internal/modules/birthdays/adapter/gateway"
	"waylight/internal/modules/birthdays/domain"
	"waylight/internal/modules/birthdays/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the birthday rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.Birthdays
	handlers *controller.Handlers
}

// New builds the module. client is the guarded integration client for the
// CardDAV address book; alerts receives the reminders (nil = none).
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, alerts kernel.AlertSink, ids kernel.IDGenerator, clock kernel.Clock, loc *time.Location) *Module {
	uc := usecase.New(usecase.Deps{Repo: gateway.NewRepository(db, cipher), Book: gateway.NewCardDAV(client), IDs: ids, Clock: clock, Location: loc, Alerts: alerts})
	return &Module{
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		uc:           uc,
		handlers:     controller.NewHandlers(uc),
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "birthdays" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"birthdays"} }

// RegisterRoutes implements modules.Module. Settings (with the CardDAV
// password) are under /settings, which guests cannot reach.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/birthdays", h.Upcoming)
	r.Protected("POST /api/birthdays", h.Create)
	r.Protected("DELETE /api/birthdays/{id}", h.Delete)
	r.Protected("POST /api/birthdays/import", h.Import)
	r.Protected("GET /api/birthdays/settings", h.Settings)
	r.Protected("PUT /api/birthdays/settings", h.SaveSettings)
}

// Jobs implements modules.Module: the address book every six hours, the
// reminder check every five minutes.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{
		{Name: "birthdays.sync", Interval: 6 * time.Hour, Run: m.uc.Sync},
		{Name: "birthdays.remind", Interval: 5 * time.Minute, Run: func(ctx context.Context) error {
			_, err := m.uc.Remind(ctx)
			return err
		}},
	}
}

// Soon lists the birthdays of the next days (for the daily briefing).
func (m *Module) Soon(ctx context.Context, days int) ([]domain.Upcoming, error) {
	return m.uc.Soon(ctx, days)
}

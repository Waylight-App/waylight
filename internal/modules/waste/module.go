package waste

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/waste/adapter/controller"
	"waylight/internal/modules/waste/adapter/gateway"
	"waylight/internal/modules/waste/adapter/smartsearch"
	"waylight/internal/modules/waste/usecase"
	"waylight/internal/shared/answers"
	"waylight/internal/shared/kernel"
)

// refreshInterval: collection calendars rarely change.
const refreshInterval = 6 * time.Hour

// Module wires the waste rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc        usecase.Waste
	reminders usecase.Reminders
	handlers  *controller.Handlers
	reminder  *controller.ReminderHandlers
}

// New builds the waste module; floating ICS times are interpreted in loc,
// which is also the time zone of the evening reminder. The reminder goes to
// alerts, the existing notification channels (#866).
func New(db *sql.DB, client *http.Client, clock kernel.Clock, alerts kernel.AlertSink, loc *time.Location) *Module {
	uc := usecase.NewWaste(gateway.NewSettingsRepository(db), gateway.NewICSSource(client, db, loc), gateway.NewParser(loc), clock)
	reminders := usecase.NewReminders(gateway.NewReminderRepository(db), uc, alerts, clock, loc)
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, uc: uc, reminders: reminders,
		handlers: controller.NewHandlers(uc), reminder: controller.NewReminderHandlers(reminders)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "waste" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"waste"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/waste/settings", m.handlers.Settings)
	r.Protected("PUT /api/waste/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/waste/upcoming", m.handlers.Upcoming)
	r.Protected("POST /api/waste/calendar", m.handlers.UploadCalendar)
	r.Protected("DELETE /api/waste/calendar", m.handlers.RemoveCalendar)
	r.Protected("GET /api/waste/reminder", m.reminder.Reminder)
	r.Protected("PUT /api/waste/reminder", m.reminder.SaveReminder)
}

// reminderInterval: the reminder job looks every minute whether its time has come.
const reminderInterval = time.Minute

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{
		{Name: "waste.refresh", Interval: refreshInterval, Run: m.uc.Refresh},
		{Name: "waste.reminder", Interval: reminderInterval, Run: m.reminders.Send},
	}
}

// Upcoming returns the next collections per type (read access for the
// briefing, ADR 0016).
func (m *Module) Upcoming(ctx context.Context) (usecase.Overview, error) { return m.uc.Upcoming(ctx) }

// AnswerSources implements answers.Provider: the waste answers of the smart
// search (ADR 0026).
func (m *Module) AnswerSources() []answers.Source {
	return []answers.Source{smartsearch.New(m.uc)}
}

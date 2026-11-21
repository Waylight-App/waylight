package speedtest

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/speedtest/adapter/controller"
	"waylight/internal/modules/speedtest/adapter/gateway"
	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/modules/speedtest/usecase"
	"waylight/internal/shared/kernel"
)

// refreshInterval: Speedtest Tracker usually tests hourly.
const refreshInterval = 10 * time.Minute

// Module wires the speedtest rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Speedtest
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
}

// New builds the speedtest module; cipher encrypts the stored token and
// sink receives threshold observations.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, sink kernel.ObservationSink, clock kernel.Clock) *Module {
	api := gateway.NewAPI(client)
	uc := usecase.NewSpeedtest(gateway.NewSettingsRepository(db, cipher), api, sink, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), specs: connectionSpecs(api, clock)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "speedtest" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"speedtest"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/speedtest/settings", m.handlers.Settings)
	r.Protected("PUT /api/speedtest/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/speedtest/results", m.handlers.Results)
}

// Jobs implements modules.Module; it does nothing until configured.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "speedtest.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}

// ConnectionStatus implements modules.StatusReporter (#592): the results
// the job keeps say whether the last refresh reached Speedtest Tracker.
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	v, err := m.uc.Results(ctx, 1)
	if err == nil && v.Error != "" {
		err = errUnreachable
	}
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

var errUnreachable = kernel.NewPublicError(kernel.ErrUnavailable, "Speedtest Tracker is not reachable.")

// compactHistoryHours is the shortest history Results accepts; the compact
// view only needs the latest measurement.
const compactHistoryHours = 1

// CompactState implements modules.CompactProvider (#638): the latest
// measurement.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	v, err := m.uc.Results(ctx, compactHistoryHours)
	if err != nil {
		return kernel.CompactState{}, err
	}
	return usecase.CompactState(v), nil
}

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

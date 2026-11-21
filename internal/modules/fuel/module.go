package fuel

import (
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/fuel/adapter/controller"
	"waylight/internal/modules/fuel/adapter/gateway"
	"waylight/internal/modules/fuel/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the fuel price rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
}

// New builds the fuel module (#53); apiURL is the Tankerkönig base URL.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, apiURL string, clock kernel.Clock) *Module {
	if apiURL == "" {
		apiURL = gateway.DefaultTankerkoenigURL
	}
	provider := gateway.NewTankerkoenig(client, apiURL)
	uc := usecase.NewFuel(gateway.NewSettingsRepository(db, cipher), provider, clock)
	return &Module{handlers: controller.NewHandlers(uc), specs: connectionSpecs(provider, apiURL)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "fuel" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"fuel"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/fuel/settings", m.handlers.Settings)
	r.Protected("PUT /api/fuel/settings", m.handlers.SaveSettings)
	r.Protected("PUT /api/fuel/favourites", m.handlers.Favourites)
	r.Protected("GET /api/fuel/overview", m.handlers.Overview)
}

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

package immich

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules/immich/domain"

	"waylight/internal/modules"
	"waylight/internal/modules/immich/adapter/controller"
	"waylight/internal/modules/immich/adapter/gateway"
	"waylight/internal/modules/immich/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the Immich rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.Immich
	handlers *controller.Handlers
	// specs are the services a named connection can reach (#610).
	specs []modules.ConnectionSpec
}

// New builds the Immich module (#85); cipher encrypts the API key.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	api := gateway.NewAPI(client)
	uc := usecase.NewImmich(gateway.NewSettingsRepository(db, cipher), api, clock)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), specs: connectionSpecs(api)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "immich" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"immich"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/immich/settings", m.handlers.Settings)
	r.Protected("PUT /api/immich/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/immich/overview", m.handlers.Overview)
	r.Protected("GET /api/immich/thumbnails/{id}", m.handlers.Thumbnail)
}

// ConnectionStatus implements modules.StatusReporter (#599): the overview
// the widget shows tells whether Immich answers.
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	_, err := m.uc.Overview(ctx)
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// Discovery implements modules.Discoverable (#590): Immich answers its
// ping without a key.
func (m *Module) Discovery() modules.DiscoverySpec {
	return modules.DiscoverySpec{
		Service: "Immich", Images: []string{"immich-server"}, Hostnames: []string{"immich-server", "immich_server", "immich.local"},
		Ports: []int{2283}, ProbePath: "/api/server/ping", ProbeContains: "pong",
	}
}

// ConnectionSpecs implements modules.Connectable (#610).
func (m *Module) ConnectionSpecs() []modules.ConnectionSpec { return m.specs }

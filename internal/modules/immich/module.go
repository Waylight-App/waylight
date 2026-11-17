package immich

import (
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/immich/adapter/controller"
	"waylight/internal/modules/immich/adapter/gateway"
	"waylight/internal/modules/immich/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the Immich rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the Immich module (#85); cipher encrypts the API key.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	uc := usecase.NewImmich(gateway.NewSettingsRepository(db, cipher), gateway.NewAPI(client), clock)
	return &Module{handlers: controller.NewHandlers(uc)}
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

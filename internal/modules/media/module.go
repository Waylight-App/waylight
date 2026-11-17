package media

import (
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/media/adapter/controller"
	"waylight/internal/modules/media/adapter/gateway"
	"waylight/internal/modules/media/domain"
	"waylight/internal/modules/media/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the media server rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the media module (#61); cipher encrypts the stored token.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock) *Module {
	providers := map[domain.Provider]usecase.MediaServerProvider{
		domain.ProviderJellyfin: gateway.NewJellyfin(client),
		domain.ProviderPlex:     gateway.NewPlex(client),
	}
	uc := usecase.NewMedia(gateway.NewSettingsRepository(db, cipher), providers, clock)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "media" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"media"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/media/settings", m.handlers.Settings)
	r.Protected("PUT /api/media/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/media/overview", m.handlers.Overview)
	r.Protected("GET /api/media/posters/{id}", m.handlers.Poster)
}

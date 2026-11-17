package space

import (
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/space/adapter/controller"
	"waylight/internal/modules/space/adapter/gateway"
	"waylight/internal/modules/space/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the space rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
}

// Config of the module; empty URLs use the public APIs.
type Config struct {
	DB         *sql.DB
	Cipher     gateway.Cipher
	Client     *http.Client
	Clock      kernel.Clock
	LaunchURL  string
	PictureURL string
}

// New builds the module (#64). Nothing is requested until a source is
// switched on in the settings and a widget shows it.
func New(c Config) *Module {
	if c.LaunchURL == "" {
		c.LaunchURL = gateway.DefaultLaunchURL
	}
	if c.PictureURL == "" {
		c.PictureURL = gateway.DefaultPictureURL
	}
	uc := usecase.New(gateway.NewSettingsRepository(c.DB, c.Cipher), gateway.NewLaunchLibrary(c.Client, c.LaunchURL),
		gateway.NewAPOD(c.Client, c.PictureURL), c.Clock)
	return &Module{Transferable: modules.Transferable{DB: c.DB, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "space" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"space"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/space/launches", m.handlers.Launches)
	r.Protected("GET /api/space/apod", m.handlers.Picture)
	r.Protected("GET /api/space/apod/image", m.handlers.Image)
	r.Protected("GET /api/space/settings", m.handlers.Settings)
	r.Protected("PUT /api/space/settings", m.handlers.SaveSettings)
}

package smarthome

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/smarthome/adapter/controller"
	"waylight/internal/modules/smarthome/adapter/gateway"
	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/modules/smarthome/usecase"
	"waylight/internal/shared/kernel"
)

// Config configures the smart home module.
type Config struct {
	DB     *sql.DB
	Cipher gateway.Cipher
	Client *http.Client
	// ActionEntities may be switched or pressed (SMARTHOME_ACTIONS_ENTITIES);
	// empty disables actions.
	ActionEntities []string
	Clock          kernel.Clock
	Audit          kernel.AuditSink
}

// Module wires the smart home rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.SmartHome
	handlers *controller.Handlers
}

// New builds the smart home module (#49) with Home Assistant as provider.
func New(cfg Config) *Module {
	uc := usecase.NewSmartHome(usecase.Deps{
		Repo: gateway.NewSettingsRepository(cfg.DB, cfg.Cipher), Provider: gateway.NewHomeAssistant(cfg.Client),
		Clock: cfg.Clock, Policy: domain.ActionPolicy{Allowed: cfg.ActionEntities}, Audit: cfg.Audit,
	})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "smarthome" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"smarthome"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/smarthome/settings", m.handlers.Settings)
	r.Protected("PUT /api/smarthome/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/smarthome/entities", m.handlers.Entities)
	r.Protected("GET /api/smarthome/states", m.handlers.States)
	r.Protected("POST /api/smarthome/entities/{id}/action", m.handlers.Act)
}

// ConnectionStatus implements modules.StatusReporter (#599): the entity
// list the widget shows tells whether Home Assistant answers.
func (m *Module) ConnectionStatus(ctx context.Context) modules.ConnectionStatus {
	_, err := m.uc.Entities(ctx)
	return modules.StatusOf(err, domain.ErrNotConfigured)
}

// Discovery implements modules.Discoverable (#590): the web app manifest of
// Home Assistant names it and needs no token.
func (m *Module) Discovery() modules.DiscoverySpec {
	return modules.DiscoverySpec{
		Service: "Home Assistant", Images: []string{"home-assistant", "homeassistant"},
		Hostnames: []string{"homeassistant.local", "homeassistant"},
		Ports:     []int{8123}, ProbePath: "/manifest.json", ProbeContains: "Home Assistant",
	}
}

package adguard

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/adguard/adapter/controller"
	"waylight/internal/modules/adguard/adapter/gateway"
	"waylight/internal/modules/adguard/domain"
	"waylight/internal/modules/adguard/usecase"
	"waylight/internal/shared/kernel"
)

const refreshInterval = time.Minute

// Module wires the DNS filter rings (AdGuard Home and Pi-hole, #83) and
// implements modules.Module. The name stays "adguard" so existing widgets,
// API clients and MODULES_DISABLED settings keep working.
type Module struct {
	modules.Base
	uc       usecase.DNSFilter
	handlers *controller.Handlers
}

// Config are the dependencies of the module.
type Config struct {
	DB *sql.DB
	// Cipher encrypts the stored passwords.
	Cipher gateway.Cipher
	Client *http.Client
	// Pausable lists the providers that may be paused (DNS_PAUSE_PROVIDERS, #177).
	Pausable []string
	Audit    kernel.AuditSink
	Clock    kernel.Clock
}

// New builds the module.
func New(c Config) *Module {
	providers := map[domain.Provider]usecase.DNSFilterProvider{
		domain.ProviderAdGuard: gateway.NewAPI(c.Client),
		domain.ProviderPiHole:  gateway.NewPiHole(c.Client),
	}
	var pausable []domain.Provider
	for _, raw := range c.Pausable {
		if p, err := domain.ParseProvider(raw); err == nil && raw != "" {
			pausable = append(pausable, p)
		}
	}
	uc := usecase.NewDNSFilter(usecase.Deps{Repo: gateway.NewSettingsRepository(c.DB, c.Cipher), Providers: providers,
		Pausable: pausable, Audit: c.Audit, Clock: c.Clock})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "adguard" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"adguard"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/adguard/settings", m.handlers.Settings)
	r.Protected("PUT /api/adguard/settings", m.handlers.SaveSettings)
	r.Protected("GET /api/adguard/stats", m.handlers.Stats)
	r.Protected("POST /api/adguard/pause", m.handlers.Pause)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "adguard.refresh", Interval: refreshInterval, Run: m.uc.Refresh}}
}

package electricity

import (
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/electricity/adapter/controller"
	"waylight/internal/modules/electricity/adapter/gateway"
	"waylight/internal/modules/electricity/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the electricity price rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the module; apiURL is the aWATTar base URL ("{tld}" = de/at).
// Nothing is requested until a widget asks for prices.
func New(client *http.Client, apiURL string, clock kernel.Clock) *Module {
	uc := usecase.New(gateway.NewAWATTar(client, apiURL), clock, time.Local)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "electricity" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"electricity"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/electricity/prices", m.handlers.Prices)
}

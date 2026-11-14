package electricity

import (
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/electricity/adapter/controller"
	"trisse-home/internal/modules/electricity/adapter/gateway"
	"trisse-home/internal/modules/electricity/usecase"
	"trisse-home/internal/shared/kernel"
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

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"electricity"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/electricity/prices", m.handlers.Prices)
}

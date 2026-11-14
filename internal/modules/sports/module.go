package sports

import (
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/sports/adapter/controller"
	"trisse-home/internal/modules/sports/adapter/gateway"
	"trisse-home/internal/modules/sports/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the sports rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the module; apiURL is the OpenLigaDB base URL. Nothing is
// requested until a widget asks.
func New(client *http.Client, apiURL string, clock kernel.Clock) *Module {
	return &Module{handlers: controller.NewHandlers(usecase.New(gateway.NewOpenLigaDB(client, apiURL), clock))}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "sports" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"sports"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/sports/leagues", m.handlers.Leagues)
	r.Protected("GET /api/sports/teams", m.handlers.Teams)
	r.Protected("GET /api/sports/overview", m.handlers.Overview)
}

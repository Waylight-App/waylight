package markets

import (
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/markets/adapter/controller"
	"trisse-home/internal/modules/markets/adapter/gateway"
	"trisse-home/internal/modules/markets/domain"
	"trisse-home/internal/modules/markets/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the markets rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the module; ecbURL and stooqURL override the sources (tests).
// Nothing is requested until a widget asks.
func New(client *http.Client, ecbURL, stooqURL string, clock kernel.Clock) *Module {
	uc := usecase.New(map[domain.Source]usecase.SeriesProvider{
		domain.SourceFX:    gateway.NewECB(client, ecbURL),
		domain.SourceStooq: gateway.NewStooq(client, stooqURL),
	}, clock)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "markets" }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"markets"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/markets/watchlist", m.handlers.Watchlist)
}

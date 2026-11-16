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
	rates    *controller.RatesHandler
}

// New builds the module; ecbURL and stooqURL override the sources (tests).
// Nothing is requested until a widget or the search asks. ratesOn switches
// the currency rates for the search on (#482); offline installations turn
// them off with CURRENCY_RATES=false.
func New(client *http.Client, ecbURL, stooqURL string, clock kernel.Clock, ratesOn bool) *Module {
	ecb := gateway.NewECB(client, ecbURL)
	uc := usecase.New(map[domain.Source]usecase.SeriesProvider{
		domain.SourceFX:    ecb,
		domain.SourceStooq: gateway.NewStooq(client, stooqURL),
	}, clock)
	return &Module{handlers: controller.NewHandlers(uc), rates: controller.NewRatesHandler(usecase.NewRates(ecb, clock, ratesOn))}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "markets" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"markets"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/markets/watchlist", m.handlers.Watchlist)
	r.Protected("GET /api/markets/rates", m.rates.Rates)
}

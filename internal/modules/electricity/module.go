package electricity

import (
	"context"
	"net/http"
	"net/url"
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
	uc       usecase.Electricity
	clock    kernel.Clock
	loc      *time.Location
	handlers *controller.Handlers
}

// New builds the module; apiURL is the aWATTar base URL ("{tld}" = de/at).
// Nothing is requested until a widget asks for prices.
func New(client *http.Client, apiURL string, clock kernel.Clock) *Module {
	uc := usecase.New(gateway.NewAWATTar(client, apiURL), clock, time.Local)
	return &Module{uc: uc, clock: clock, loc: time.Local, handlers: controller.NewHandlers(uc)}
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

// CompactStateFor implements modules.CompactParamProvider (#641): the
// current price of the widget's zone, chosen with ?zone= (DE-LU or AT).
func (m *Module) CompactStateFor(ctx context.Context, params url.Values) (kernel.CompactState, error) {
	p, err := m.uc.Prices(ctx, params.Get("zone"))
	if err != nil {
		return kernel.CompactState{}, err
	}
	return usecase.CompactState(p, m.clock.Now(), m.loc), nil
}

// CompactParams implements modules.CompactParamProvider.
func (m *Module) CompactParams() []string { return []string{"zone"} }

package transit

import (
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/transit/adapter/controller"
	"waylight/internal/modules/transit/adapter/gateway"
	"waylight/internal/modules/transit/adapter/smartsearch"
	"waylight/internal/modules/transit/usecase"
	"waylight/internal/shared/answers"
	"waylight/internal/shared/kernel"
)

// Module wires the transit rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
	uc       usecase.Transit
}

// New builds the transit module; apiURL points to a MOTIS instance.
func New(client *http.Client, apiURL string, clock kernel.Clock) *Module {
	if apiURL == "" {
		apiURL = gateway.DefaultURL
	}
	uc := usecase.NewTransit(gateway.NewMOTIS(client, apiURL), clock)
	return &Module{handlers: controller.NewHandlers(uc), uc: uc}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "transit" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"transit"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/transit/stops", m.handlers.Stops)
	r.Protected("GET /api/transit/departures", m.handlers.Departures)
}

// AnswerSources implements answers.Provider: the departures of the smart
// search (ADR 0026).
func (m *Module) AnswerSources() []answers.Source {
	return []answers.Source{smartsearch.New(m.uc)}
}

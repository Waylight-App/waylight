package attention

import (
	"context"

	"waylight/internal/modules"
	"waylight/internal/modules/attention/adapter/controller"
	"waylight/internal/modules/attention/domain"
	"waylight/internal/modules/attention/usecase"
	"waylight/internal/shared/kernel"
)

// Widgets lists the widgets on the account's own tabs (the dashboard).
type Widgets func(ctx context.Context) ([]usecase.Widget, error)

// Module is the attention list (#599). It has no data of its own: it asks
// the other modules through modules.StatusReporter and modules.WidgetChecker.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New collects the sources from the registered modules; build it after
// them, like the smart search.
func New(all []modules.Module, widgets Widgets, mayRead func(ctx context.Context, module string) bool, clock kernel.Clock) *Module {
	var sources []usecase.Source
	for _, m := range all {
		src := usecase.Source{Module: m.Name(), Kinds: m.WidgetKinds(), Instance: m.DataKind() == modules.InstanceData}
		if r, ok := m.(modules.StatusReporter); ok {
			src.Status = func(ctx context.Context) domain.Status { return status(r.ConnectionStatus(ctx)) }
		}
		if c, ok := m.(modules.WidgetChecker); ok {
			src.Check = func(kind, config string) *domain.Need {
				need := c.CheckWidget(kind, config)
				if need == nil {
					return nil
				}
				return &domain.Need{Reason: message(need.Reason), Field: need.Field}
			}
		}
		if src.Status != nil || src.Check != nil {
			sources = append(sources, src)
		}
	}
	uc := usecase.New(usecase.Deps{Widgets: widgets, MayRead: mayRead, Sources: sources, Clock: clock})
	return &Module{handlers: controller.NewHandlers(uc)}
}

func status(s modules.ConnectionStatus) domain.Status {
	return domain.Status{State: domain.State(s.State), Reason: message(s.Reason), Action: domain.Action(s.Action)}
}

func message(e *kernel.PublicError) domain.Message {
	if e == nil {
		return domain.Message{}
	}
	return domain.Message{Text: e.Message, Template: e.Template, Params: e.Params}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "attention" }

// DataKind says the list belongs to the asking account; it stores nothing.
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/attention", m.handlers.List)
	r.Protected("GET /api/attention/connections", m.handlers.Connections)
}

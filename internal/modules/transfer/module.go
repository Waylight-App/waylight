package transfer

import (
	"context"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/transfer/adapter/controller"
	"trisse-home/internal/modules/transfer/adapter/gateway"
	"trisse-home/internal/modules/transfer/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the transfer rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
	config   *controller.ConfigHandlers
	foreign  *controller.ForeignHandlers
	uc       usecase.Config
}

// Deps are the ports of the transfer module. Portables and Configurables
// are called per request (usually the registry's methods) so modules
// registered later are included. Env looks up ${NAME} in config files.
type Deps struct {
	Portables     func() map[string]modules.Portable
	Configurables func() map[string]modules.Configurable
	Env           func(string) (string, bool)
	// WidgetKinds reports whether a widget kind is available (#77).
	WidgetKinds func(string) bool
	Clock       kernel.Clock
	Audit       kernel.AuditSink
}

// New builds the transfer module.
func New(d Deps) *Module {
	sources := func() map[string]usecase.ConfigSource {
		out := map[string]usecase.ConfigSource{}
		for name, c := range d.Configurables() {
			out[name] = c
		}
		return out
	}
	managed := &usecase.Managed{}
	cfg := usecase.NewConfig(sources, gateway.YAMLCodec{}, d.Env, d.Audit, managed)
	return &Module{
		handlers: controller.NewHandlers(usecase.NewTransfer(gateway.Sections(d.Portables), d.Clock, d.Audit, managed)),
		config:   controller.NewConfigHandlers(cfg),
		foreign: controller.NewForeignHandlers(usecase.NewImporter(map[string]usecase.ImportSource{
			"homepage": gateway.Homepage{}, "homarr": gateway.Homarr{},
		}, cfg, gateway.YAMLCodec{}, d.WidgetKinds)),
		uc: cfg,
	}
}

// ApplyManagedFile applies CONFIG_FILE at start-up (called from the
// composition root). On success the sections in the file are managed:
// imports skip them and Managed reports them.
func (m *Module) ApplyManagedFile(ctx context.Context, data []byte) ([]string, error) {
	return m.uc.ApplyManaged(ctx, data)
}

// Managed lists the sections CONFIG_FILE owns.
func (m *Module) Managed() []string { return m.uc.Managed() }

// Name implements modules.Module.
func (m *Module) Name() string { return "transfer" }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/transfer/export", m.handlers.Export)
	r.Protected("POST /api/transfer/import", m.handlers.Import)
	r.Protected("GET /api/transfer/config", m.config.Export)
	r.Protected("POST /api/transfer/config", m.config.Apply)
	r.Protected("GET /api/transfer/config/status", m.config.Status)
	r.Protected("POST /api/transfer/import/{format}", m.foreign.Import)
}

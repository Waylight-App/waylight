package connections

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/connections/adapter/configfile"
	"waylight/internal/modules/connections/adapter/controller"
	"waylight/internal/modules/connections/adapter/gateway"
	"waylight/internal/modules/connections/usecase"
	"waylight/internal/shared/kernel"
)

// Widgets lists widgets with the connection their settings name, so a
// connection can say which widgets use it.
type Widgets func(ctx context.Context) ([]usecase.Widget, error)

// Module wires the Connections rings and implements modules.Module and
// modules.Configurable. It is not Portable: a data export never carries
// credentials, and connections without them are of no use.
type Module struct {
	modules.Base
	handlers *controller.Handlers
	config   usecase.Config
}

// New builds the connections module (#610); cipher encrypts credentials.
func New(db *sql.DB, cipher gateway.Cipher, widgets Widgets, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	repo := gateway.NewRepository(db, cipher)
	uc := usecase.New(usecase.Deps{Repo: repo, Widgets: widgets, Clock: clock, IDs: ids})
	return &Module{handlers: controller.NewHandlers(uc), config: usecase.NewConfig(repo, clock, ids)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "connections" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind says connections belong to the instance and its admins.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/connections", m.handlers.List)
	r.Protected("POST /api/connections", m.handlers.Add)
	r.Protected("PUT /api/connections/{id}", m.handlers.Update)
	r.Protected("PUT /api/connections/{id}/credential", m.handlers.ReplaceCredential)
	r.Protected("DELETE /api/connections/{id}", m.handlers.Remove)
}

// ConfigSection implements modules.Configurable.
func (m *Module) ConfigSection() string { return "connections" }

// ExportConfig implements modules.Configurable.
func (m *Module) ExportConfig(ctx context.Context) (any, error) {
	list, err := m.config.Export(ctx)
	if err != nil {
		return nil, err
	}
	return configfile.FromList(list), nil
}

// ConfigRoutes implements modules.Configurable. Replacing a credential is
// not listed: the file names connections, the credential is entered here.
func (m *Module) ConfigRoutes() []string {
	return []string{
		"POST /api/connections",
		"PUT /api/connections/{id}",
		"DELETE /api/connections/{id}",
	}
}

// NewConfig implements modules.Configurable.
func (m *Module) NewConfig() any { return &configfile.Section{} }

// ApplyConfig implements modules.Configurable.
func (m *Module) ApplyConfig(ctx context.Context, section any) error {
	s, ok := section.(*configfile.Section)
	if !ok {
		return modules.ErrConfigType
	}
	return m.config.Apply(ctx, s.List())
}

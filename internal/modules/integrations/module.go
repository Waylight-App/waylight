package integrations

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/integrations/adapter/configfile"
	"waylight/internal/modules/integrations/adapter/controller"
	"waylight/internal/modules/integrations/adapter/gateway"
	"waylight/internal/modules/integrations/domain"
	"waylight/internal/modules/integrations/usecase"
	"waylight/internal/shared/kernel"
)

// Widgets lists widgets with the connection their settings name, so a
// connection can say which widgets use it.
type Widgets func(ctx context.Context) ([]usecase.Widget, error)

// Module wires the Integrations rings and implements modules.Module and
// modules.Configurable. It is not Portable: a data export never carries
// credentials, and connections without them are of no use.
type Module struct {
	modules.Base
	handlers *controller.Handlers
	config   usecase.Config
}

// New builds the integrations module (#610); cipher encrypts credentials,
// specs are the services of the enabled modules (modules.ConnectionSpecsOf).
func New(db *sql.DB, cipher gateway.Cipher, specs []modules.ConnectionSpec, widgets Widgets, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	repo := gateway.NewRepository(db, cipher)
	uc := usecase.New(usecase.Deps{Repo: repo, Widgets: widgets, Services: services(specs), Clock: clock, IDs: ids})
	return &Module{handlers: controller.NewHandlers(uc), config: usecase.NewConfig(repo, clock, ids)}
}

// services converts the module specs into the ports of the use cases, so
// the inner rings do not depend on the module contract.
func services(specs []modules.ConnectionSpec) []usecase.Service {
	out := make([]usecase.Service, 0, len(specs))
	for _, s := range specs {
		spec := domain.Spec{Service: s.Service, Title: s.Title, FixedURL: s.FixedURL}
		for _, f := range s.Fields {
			spec.Fields = append(spec.Fields, domain.Field{Name: f.Name, Secret: f.Secret, Optional: f.Optional})
		}
		out = append(out, usecase.Service{Spec: spec, Check: s.Check})
	}
	return out
}

// Name implements modules.Module.
func (m *Module) Name() string { return "integrations" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind says connections belong to the instance and its admins.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/integrations/services", m.handlers.Services)
	r.Protected("GET /api/integrations/connections", m.handlers.List)
	r.Protected("POST /api/integrations/connections/{id}/check", m.handlers.Check)
	r.Protected("POST /api/integrations/connections", m.handlers.Add)
	r.Protected("PUT /api/integrations/connections/{id}", m.handlers.Update)
	r.Protected("PUT /api/integrations/connections/{id}/credential", m.handlers.ReplaceCredential)
	r.Protected("DELETE /api/integrations/connections/{id}", m.handlers.Remove)
}

// ConfigSection implements modules.Configurable.
func (m *Module) ConfigSection() string { return "integrations" }

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
		"POST /api/integrations/connections",
		"PUT /api/integrations/connections/{id}",
		"DELETE /api/integrations/connections/{id}",
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

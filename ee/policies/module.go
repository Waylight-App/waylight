// Package policies holds the admin policies of the ee/ features (#304):
// required second factor, minimum password length, maximum session age and
// the modules members may use. The auth module applies them.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic.
package policies

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/ee/licence"
	"waylight/ee/policies/adapter/controller"
	"waylight/ee/policies/adapter/gateway"
	"waylight/ee/policies/usecase"
	"waylight/internal/modules"
	"waylight/internal/modules/auth"
)

// Module wires the rings.
type Module struct {
	modules.Base
	svc       *usecase.Service
	handlers  *controller.Handlers
	licensing licence.Licensing
}

// New builds the module and installs it as the auth module's policy source
// and the registry's member filter.
func New(db *sql.DB, a *auth.Module, registry *modules.Registry, licensing licence.Licensing) *Module {
	svc := usecase.New(usecase.Deps{Store: gateway.NewStore(db), Licensing: licensing, Audit: a})
	m := &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: licensing}
	a.SetPolicySource(m)
	registry.SetMemberModules(func(ctx context.Context, module string) bool { return m.Policy(ctx).MemberMayUse(module) })
	return m
}

// Policy implements auth.PolicySource.
func (m *Module) Policy(ctx context.Context) auth.Policy {
	set := m.svc.Current(ctx)
	return auth.Policy{RequireSecondFactor: set.RequireSecondFactor, MinLength: set.MinLength,
		MaxSessionAge: set.MaxSessionAge(), MemberModules: set.MemberModules}
}

// Access implements modules.Gated: changing policies needs the feature.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeaturePolicies)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "policies" }

// DataKind implements modules.Module.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module; admins only (use cases).
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/policies", m.handlers.Get)
	r.Protected("PUT /api/policies", m.handlers.Save)
}

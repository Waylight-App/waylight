// Package scim provisions members and groups from a directory over SCIM 2.0,
// part of directory sync in the ee/ features (#396).
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic.
package scim

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/ee/licence"
	"trisse-home/ee/scim/adapter/controller"
	"trisse-home/ee/scim/adapter/gateway"
	"trisse-home/ee/scim/usecase"
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
)

// Module wires the rings.
type Module struct {
	modules.Base
	handlers  *controller.Handlers
	licensing licence.Licensing
}

// New builds the module; token is SCIM_TOKEN (empty answers every request 401).
func New(db *sql.DB, a *auth.Module, licensing licence.Licensing, token string) *Module {
	svc := usecase.New(usecase.Deps{Accounts: a.Provisioning(), Store: gateway.NewStore(db), Token: token})
	return &Module{handlers: controller.NewHandlers(svc), licensing: licensing}
}

// Access implements modules.Gated: SCIM needs a key with directory-sync.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureDirectorySync)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "scim" }

// DataKind implements modules.Module.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module. The routes are public for the
// session guard: the identity provider authenticates with the SCIM token.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Public("GET /scim/v2/ServiceProviderConfig", h.ServiceProviderConfig)
	r.Public("GET /scim/v2/Users", h.ListUsers)
	r.Public("POST /scim/v2/Users", h.CreateUser)
	r.Public("GET /scim/v2/Users/{id}", h.GetUser)
	r.Public("PUT /scim/v2/Users/{id}", h.ReplaceUser)
	r.Public("PATCH /scim/v2/Users/{id}", h.PatchUser)
	r.Public("DELETE /scim/v2/Users/{id}", h.DeleteUser)
	r.Public("GET /scim/v2/Groups", h.ListGroups)
	r.Public("POST /scim/v2/Groups", h.CreateGroup)
	r.Public("GET /scim/v2/Groups/{id}", h.GetGroup)
	r.Public("PUT /scim/v2/Groups/{id}", h.ReplaceGroup)
	r.Public("PATCH /scim/v2/Groups/{id}", h.PatchGroup)
	r.Public("DELETE /scim/v2/Groups/{id}", h.DeleteGroup)
}

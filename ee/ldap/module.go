// Package ldap signs members in with their LDAP or Active Directory password,
// part of directory sync in the ee/ features (#396).
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module; it contains no business logic.
package ldap

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"

	"trisse-home/ee/ldap/adapter/gateway"
	"trisse-home/ee/ldap/domain"
	"trisse-home/ee/ldap/usecase"
	"trisse-home/ee/licence"
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
)

// Config of the directory; see gateway.Config.
type Config = gateway.Config

// Policy decides who becomes a member; see domain.Policy.
type Policy = domain.Policy

// Module wires the rings.
type Module struct{ modules.Base }

// New builds the module and hands its authenticator to the auth module.
// Without cfg.URL there is no directory, and names are checked locally.
func New(db *sql.DB, a *auth.Module, licensing licence.Licensing, cfg Config, policy Policy) *Module {
	if cfg.URL == "" {
		return &Module{}
	}
	authn := usecase.New(usecase.Deps{
		Directory: gateway.NewDirectory(cfg), Links: gateway.NewLinks(db), Accounts: a.Provisioning(),
		Licensing: licensing, Policy: policy,
	})
	a.SetDirectoryAuthenticator(authenticator{authn})
	return &Module{}
}

// authenticator maps the use case's "check locally" to the auth module's.
type authenticator struct{ a *usecase.Authenticator }

func (x authenticator) Authenticate(ctx context.Context, name, password string) (string, error) {
	id, err := x.a.Authenticate(ctx, name, password)
	if errors.Is(err, usecase.ErrNotInDirectory) {
		return "", auth.ErrNotInDirectory
	}
	return id, err
}

// Name implements modules.Module.
func (m *Module) Name() string { return "ldap" }

// DataKind implements modules.Module.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

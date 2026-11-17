// Package sso is single sign-on for members of the ee/ features (#302): the
// identity provider decides who becomes a member and which groups they are
// in. The owner's single sign-on stays in the core (ADR 0013).
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module; it contains no business logic.
package sso

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/ee/licence"
	"trisse-home/ee/sso/adapter/gateway"
	"trisse-home/ee/sso/domain"
	"trisse-home/ee/sso/usecase"
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/shared/kernel"
)

// Policy decides who becomes a member; see domain.Policy.
type Policy = domain.Policy

// Module wires the rings and implements modules.Module.
type Module struct {
	modules.Base
	resolver *usecase.Resolver
}

// New builds the module and hands its resolver to the auth module, which asks
// it about everyone who signs in through OpenID Connect and is not the owner.
func New(db *sql.DB, clock kernel.Clock, a *auth.Module, licensing licence.Licensing, policy Policy) *Module {
	m := &Module{resolver: usecase.New(usecase.Deps{
		Links: gateway.NewLinks(db, clock), Accounts: a.Provisioning(), Licensing: licensing, Policy: policy,
	})}
	a.SetExternalAccounts(externalAccounts{m.resolver})
	return m
}

// externalAccounts adapts the resolver to the auth module's port.
type externalAccounts struct{ r *usecase.Resolver }

// SSOLicensed implements the auth module's SSOLicence (#527): linked
// sign-ins of members and guests count with any licence level, new links
// need full use.
func (e externalAccounts) SSOLicensed(ctx context.Context) (signIn, link bool) {
	return e.r.Licensed(ctx)
}

func (e externalAccounts) Resolve(ctx context.Context, id auth.ExternalIdentity) (string, error) {
	return e.r.Resolve(ctx, domain.Identity{Subject: id.Subject, Email: id.Email, EmailVerified: id.EmailVerified,
		Name: id.Name, Username: id.Username, Groups: id.Groups})
}

// Name implements modules.Module.
func (m *Module) Name() string { return "sso" }

// DataKind implements modules.Module: the links belong to the instance.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

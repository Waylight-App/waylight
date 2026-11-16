// Package connections lets members connect their own accounts at external
// services, for the team integrations of the ee/ features (#326). The
// integrations ask AccessToken for a valid token of a member.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic.
package connections

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"

	"trisse-home/ee/connections/adapter/controller"
	"trisse-home/ee/connections/adapter/gateway"
	"trisse-home/ee/connections/usecase"
	"trisse-home/ee/licence"
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/auth"
	"trisse-home/internal/shared/kernel"
)

// Module wires the rings.
type Module struct {
	modules.Base
	svc       *usecase.Service
	handlers  *controller.Handlers
	licensing licence.Licensing
}

// Deps of the module. Audit receives connect and disconnect for the audit log.
type Deps struct {
	DB        *sql.DB
	Cipher    gateway.Cipher
	Client    *http.Client
	Clock     kernel.Clock
	PublicURL *url.URL
	Auth      *auth.Module
	Licensing licence.Licensing
	Audit     func(ctx context.Context, action, service string)
	Log       *slog.Logger
	// OAuth replaces the adapter, for tests.
	OAuth usecase.OAuth
}

// New builds the module and removes a deleted account's tokens with it.
func New(d Deps) *Module {
	store := gateway.NewStore(d.DB, d.Cipher)
	o := d.OAuth
	if o == nil {
		o = gateway.NewOAuth(d.Client, d.Clock.Now, d.PublicURL)
	}
	svc := usecase.New(usecase.Deps{Tokens: store, Clients: store.Clients(), OAuth: o, Clock: d.Clock,
		Licensing: d.Licensing, Audit: d.Audit})
	d.Auth.OnAccountDeleted(func(ctx context.Context, id string) {
		if err := svc.AccountDeleted(ctx, id); err != nil && d.Log != nil {
			d.Log.Error("connected accounts: could not delete the tokens of a deleted account", "err", err)
		}
	})
	return &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: d.Licensing}
}

// AccessToken returns a valid access token of a member at a service, for the
// team integrations; see usecase.Service.AccessToken.
func (m *Module) AccessToken(ctx context.Context, accountID, service string) (string, error) {
	return m.svc.AccessToken(ctx, accountID, service)
}

// Access implements modules.Gated.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureConnectedAccounts)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "connections" }

// DataKind implements modules.Module: each connection belongs to one account.
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/connections", h.List)
	r.Protected("GET /api/connections/callback", h.Callback)
	r.Protected("POST /api/connections/{service}", h.Start)
	r.Protected("DELETE /api/connections/{service}", h.Disconnect)
	r.Protected("GET /api/connections/clients", h.Clients)
	r.Protected("PUT /api/connections/clients/{service}", h.SaveClient)
}

//go:build ee

package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"trisse-home/ee/connections"
	"trisse-home/ee/connections/adapter/gateway"
	"trisse-home/internal/modules"
	"trisse-home/internal/platform/oauth"
	"trisse-home/internal/shared/kernel"
)

// connectionEndpoints points services at a fake provider in tests; nil in
// production.
var connectionEndpoints func(service string) (oauth.Endpoint, bool)

var (
	connectionsMu   sync.Mutex
	connectionsMods = map[*modules.Registry]*connections.Module{}
)

// Connected accounts (#326), part of the ee/ features: members connect their
// own accounts at services for the team integrations.
func init() {
	registerModule("connections", func(d moduleDeps) modules.Module { return connectionsModule(d) })
}

// connectionsModule returns the connected accounts of the server being built,
// creating them on first use; the team integrations take their tokens from it
// whatever the build order (like licenceModule).
func connectionsModule(d moduleDeps) *connections.Module {
	connectionsMu.Lock()
	defer connectionsMu.Unlock()
	if m, ok := connectionsMods[d.registry]; ok {
		return m
	}
	m := newConnections(d)
	connectionsMods[d.registry] = m
	return m
}

func newConnections(d moduleDeps) *connections.Module {
	var o *gateway.OAuth
	if connectionEndpoints != nil {
		o = gateway.NewOAuth(d.out.integration, time.Now, d.cfg.PublicURL).WithEndpoints(connectionEndpoints)
	}
	deps := connections.Deps{
		DB: d.db, Cipher: d.out.secrets, Client: d.out.integration, Clock: d.clock, PublicURL: d.cfg.PublicURL,
		Auth: d.auth, Licensing: licenceModule(d).Licensing(), Log: slog.Default(),
		Audit: func(ctx context.Context, action, service string) {
			// Personal settings are not in the audit log by default; connecting
			// an account to the organisation's services is, so it is reported here.
			if changeSink != nil {
				p := kernel.PrincipalFrom(ctx)
				changeSink.RecordChange(ctx, modules.Change{AccountID: p.AccountID, Role: p.Role, Module: "connections",
					Route: action + " " + service, ObjectID: service})
			}
		},
	}
	if o != nil {
		deps.OAuth = o
	}
	return connections.New(deps)
}

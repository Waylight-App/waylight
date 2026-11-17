// Package slackcmd answers a Slack slash command such as /home vpn with
// results from the start page, for the member behind it (#333, part of the
// chat apps in the ee/ features). It connects out to Slack over Socket Mode
// (ADR 0023), so the instance needs no public address.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module; it contains no business logic and stores nothing.
package slackcmd

import (
	"context"
	"net/http"
	"time"

	"waylight/ee/slackcmd/adapter/slackapi"
	"waylight/ee/slackcmd/adapter/socket"
	"waylight/ee/slackcmd/usecase"
	"waylight/internal/modules"
)

// Module runs the connection as a background job.
type Module struct {
	modules.Base
	svc      *usecase.Service
	api      slackapi.Client
	appToken string
}

// Deps are what the module needs from the server.
type Deps struct {
	// AppToken is the workspace app's app-level token (xapp-…, scope
	// connections:write); empty switches the command off.
	AppToken    string
	Client      *http.Client
	API         string // https://slack.com/api; tests use a fake
	Connections usecase.Connections
	Finder      usecase.Finder
	Licensing   usecase.Licensing
	SettingsURL string
}

// New builds the module.
func New(d Deps) *Module {
	api := slackapi.Client{HTTP: d.Client, API: d.API}
	svc := usecase.New(usecase.Deps{Connections: d.Connections, Slack: api, Finder: d.Finder, Licensing: d.Licensing,
		Now: time.Now, SettingsURL: d.SettingsURL})
	return &Module{svc: svc, api: api, appToken: d.AppToken}
}

// connectionLifetime is how long one job run keeps the connection; the
// scheduler starts the next run right away, so it is renewed hourly.
const connectionLifetime = time.Hour

// Jobs implements modules.Module: one long-running job while an app token
// is set.
func (m *Module) Jobs() []modules.Job {
	if m.appToken == "" {
		return nil
	}
	return []modules.Job{{Name: "slackcmd.socket", Interval: connectionLifetime, Run: func(ctx context.Context) error {
		return socket.Run(ctx,
			func(ctx context.Context) (string, error) { return m.api.OpenSocket(ctx, m.appToken) },
			func(ctx context.Context, c socket.Command) string {
				return m.svc.Handle(ctx, usecase.Command{TeamID: c.TeamID, UserID: c.UserID, Command: c.Command, Text: c.Text})
			},
			m.svc.Warm)
	}}}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "slackcmd" }

// DataKind implements modules.Module: nothing is stored.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

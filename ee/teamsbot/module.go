// Package teamsbot answers searches typed into the Microsoft Teams command
// box (#334, part of the chat apps in the ee/ features). It is an opt-in
// (ADR 0023): only with the organisation's own bot registration does the
// route exist, and Microsoft must be able to reach it.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module; it contains no business logic and stores nothing.
package teamsbot

import (
	"context"
	"net/http"
	"time"

	"trisse-home/ee/teamsbot/adapter/botauth"
	"trisse-home/ee/teamsbot/adapter/controller"
	"trisse-home/ee/teamsbot/adapter/graph"
	"trisse-home/ee/teamsbot/usecase"
	"trisse-home/internal/modules"
)

// Module wires the rings.
type Module struct {
	modules.Base
	svc      *usecase.Service
	handlers *controller.Handlers
	enabled  bool
}

// Deps are what the module needs from the server.
type Deps struct {
	// AppID is the organisation's bot app ID (TEAMS_BOT_APP_ID); empty
	// leaves the route out.
	AppID       string
	Client      *http.Client
	MetadataURL string // Bot Framework OpenID metadata; tests use a fake
	GraphAPI    string
	Connections usecase.Connections
	Finder      usecase.Finder
	Licensing   usecase.Licensing
}

// New builds the module.
func New(d Deps) *Module {
	svc := usecase.New(usecase.Deps{Connections: d.Connections, Directory: graph.Client{HTTP: d.Client, API: d.GraphAPI},
		Finder: d.Finder, Licensing: d.Licensing, Now: time.Now})
	v := &botauth.Verifier{HTTP: d.Client, MetadataURL: d.MetadataURL, AppID: d.AppID, Now: time.Now}
	return &Module{svc: svc, handlers: controller.NewHandlers(v, svc), enabled: d.AppID != ""}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "teamsbot" }

// DataKind implements modules.Module: nothing is stored.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// RegisterRoutes implements modules.Module. The route is public: Microsoft
// signs in with its own token, which the handler checks first.
func (m *Module) RegisterRoutes(r modules.Router) {
	if m.enabled {
		r.Public("POST /api/teams/messages", m.handlers.Messages)
	}
}

// warmInterval refreshes the map of Teams users now and then.
const warmInterval = 15 * time.Minute

// Jobs implements modules.Module: keeps the member map warm so the first
// query of a person needs no Graph call.
func (m *Module) Jobs() []modules.Job {
	if !m.enabled {
		return nil
	}
	return []modules.Job{{Name: "teamsbot.members", Interval: warmInterval, Run: func(ctx context.Context) error {
		m.svc.Warm(ctx)
		return nil
	}}}
}

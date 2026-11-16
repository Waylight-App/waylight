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

	licence "trisse-home/ee/licence/domain"

	"trisse-home/ee/teamsbot/adapter/botauth"
	"trisse-home/ee/teamsbot/adapter/controller"
	"trisse-home/ee/teamsbot/adapter/graph"
	"trisse-home/ee/teamsbot/adapter/tab"
	"trisse-home/ee/teamsbot/usecase"
	"trisse-home/internal/modules"
)

// Module wires the rings.
type Module struct {
	modules.Base
	svc      *usecase.Service
	handlers *controller.Handlers
	tab      tab.Handlers
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
	// Name is the instance's name for the tab page; PublicURL is what its
	// button opens (empty: no tab).
	Name      func(ctx context.Context) string
	PublicURL string
}

// New builds the module.
func New(d Deps) *Module {
	svc := usecase.New(usecase.Deps{Connections: d.Connections, Directory: graph.Client{HTTP: d.Client, API: d.GraphAPI},
		Finder: d.Finder, Licensing: d.Licensing, Now: time.Now})
	v := &botauth.Verifier{HTTP: d.Client, MetadataURL: d.MetadataURL, AppID: d.AppID, Now: time.Now}
	t := tab.Handlers{Name: d.Name, Allowed: func(ctx context.Context) bool {
		return d.Licensing.Level(ctx, licence.FeatureChatApps) != licence.LevelNone
	}}
	if d.PublicURL != "" {
		t.StartURL = tab.StartURL(d.PublicURL)
	}
	return &Module{svc: svc, handlers: controller.NewHandlers(v, svc), tab: t, enabled: d.AppID != ""}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "teamsbot" }

// DataKind implements modules.Module: nothing is stored.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// RegisterRoutes implements modules.Module. The tab page is public: Teams
// shows it before anyone signs in, and it only links to the start page. The
// bot route is public too: Microsoft signs in with its own token, which the
// handler checks first.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Public("GET /teams/tab", m.tab.Page)
	r.Public("GET /teams/tab.js", m.tab.Script)
	r.Public("GET /teams/tab.css", m.tab.Style)
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

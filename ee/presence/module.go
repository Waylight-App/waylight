// Package presence shows who is in the office, at home or away today
// (#346, part of the ee/ features). Everyone sets their own status once a
// day; planned absences from #299 win, and nothing is ever inferred.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic.
package presence

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"waylight/ee/licence"
	"waylight/ee/presence/adapter/controller"
	"waylight/ee/presence/adapter/gateway"
	"waylight/ee/presence/usecase"
	"waylight/internal/modules"
	"waylight/internal/shared/kernel"
)

// pruneInterval: statuses of earlier days go within an hour after midnight.
const pruneInterval = time.Hour

// Module wires the rings.
type Module struct {
	modules.Base
	svc       *usecase.Service
	handlers  *controller.Handlers
	licensing licence.Licensing
}

// Deps are what the module needs from the server.
type Deps struct {
	DB        *sql.DB
	Licensing licence.Licensing
	Directory usecase.Directory
	// Absences is nil without the absences module.
	Absences usecase.Absences
	Clock    kernel.Clock
}

// New builds the module.
func New(d Deps) *Module {
	svc := usecase.New(usecase.Deps{Store: gateway.NewStore(d.DB), Directory: d.Directory, Absences: d.Absences, Clock: d.Clock})
	return &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: d.Licensing}
}

// ForgetAccount removes a deleted account's status.
func (m *Module) ForgetAccount(ctx context.Context, accountID string) error {
	return m.svc.ForgetAccount(ctx, accountID)
}

// Access implements modules.Gated: readable with an expired key, no new
// statuses.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeaturePresence)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "presence" }

// DataKind implements modules.Module. Members set their own status; the
// use cases decide who may change what, as with personal settings.
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"presence"} }

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "presence.prune", Interval: pruneInterval, Run: m.svc.Prune}}
}

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/presence", h.Board)
	r.Protected("PUT /api/presence/me", h.SetMine)
	r.Protected("GET /api/presence/settings", h.Settings)
	r.Protected("PUT /api/presence/settings", h.SaveSettings)
}

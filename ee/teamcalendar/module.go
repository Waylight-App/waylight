// Package teamcalendar shares calendar subscriptions with a group or
// everyone (#299, "team calendar", part of the ee/ features): the calendar
// module shows calendars others shared only while this module's licence
// feature is on.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic and
// stores nothing (the visibility lives with the calendar's subscriptions).
package teamcalendar

import (
	"context"

	"waylight/ee/licence"
	"waylight/ee/teamcalendar/adapter/controller"
	"waylight/ee/teamcalendar/usecase"
	"waylight/internal/modules"
)

// Module wires the rings.
type Module struct {
	modules.Base
	handlers  *controller.Handlers
	licensing licence.Licensing
}

// New builds the module.
func New(sharing usecase.Sharing, licensing licence.Licensing) *Module {
	return &Module{handlers: controller.NewHandlers(usecase.New(sharing)), licensing: licensing}
}

// Access implements modules.Gated (licence feature absences: "absences and
// team calendar").
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureAbsences)
}

// SharedOn tells the calendar whether to show calendars others shared: as
// long as the licence lets them be read.
func (m *Module) SharedOn(ctx context.Context) bool { return m.Access(ctx) != modules.AccessNone }

// Name implements modules.Module.
func (m *Module) Name() string { return "teamcalendar" }

// DataKind implements modules.Module: members share their own calendars,
// so their writes pass the central guard; the use cases check whose.
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/teamcalendar/subscriptions", m.handlers.List)
	r.Protected("PUT /api/teamcalendar/subscriptions/{id}", m.handlers.Share)
	r.Protected("GET /api/teamcalendar/accounts", m.handlers.Accounts)
	r.Protected("PUT /api/teamcalendar/accounts/{id}", m.handlers.ShareAccount)
}

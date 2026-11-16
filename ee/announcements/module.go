// Package announcements holds short notices for the whole instance (#297,
// part of the ee/ features): admins or a chosen group write them, everyone
// reads them in a widget, members and admins mark them read.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module and modules.Gated; it contains no business logic.
package announcements

import (
	"context"
	"database/sql"
	"io/fs"
	"net/url"
	"strings"

	"trisse-home/ee/announcements/adapter/controller"
	"trisse-home/ee/announcements/adapter/gateway"
	"trisse-home/ee/announcements/usecase"
	"trisse-home/ee/licence"
	licencedomain "trisse-home/ee/licence/domain"
	"trisse-home/internal/modules"
	"trisse-home/internal/shared/kernel"
)

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
	Names     usecase.Names
	Alerts    kernel.AlertSink
	IDs       kernel.IDGenerator
	Clock     kernel.Clock
	Audit     func(ctx context.Context, action, id string)
	// Chat posts to Slack and Teams channels (#332); nil without
	// notifications.
	Chat usecase.Chat
	// PublicURL builds the link back to an announcement; nil adds none.
	PublicURL *url.URL
}

// New builds the module.
func New(d Deps) *Module {
	deps := usecase.Deps{Store: gateway.NewStore(d.DB), Names: d.Names, Alerts: d.Alerts, IDs: d.IDs, Clock: d.Clock,
		Audit: d.Audit, Chat: d.Chat,
		MayPost: func(ctx context.Context) bool {
			return d.Licensing != nil && d.Licensing.Level(ctx, licence.FeatureChatApps) == licencedomain.LevelFull
		}}
	if d.PublicURL != nil {
		base := strings.TrimSuffix(d.PublicURL.String(), "/")
		deps.Link = func(id string) string { return base + "/?announcement=" + url.QueryEscape(id) }
	}
	svc := usecase.New(deps)
	return &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: d.Licensing}
}

// ForgetAccount drops a deleted account's read marks.
func (m *Module) ForgetAccount(ctx context.Context, accountID string) error {
	return m.svc.ForgetAccount(ctx, accountID)
}

// Access implements modules.Gated: with an expired key notices stay
// readable, but nothing can be written.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureAnnouncements)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "announcements" }

// WidgetKinds implements modules.Module: the widget of the ee/ plugin
// (ADR 0021); it disappears from the picker with the licence.
func (m *Module) WidgetKinds() []string { return []string{"announcements"} }

// DataKind implements modules.Module. Members write their read marks and,
// in the writer group, notices; who may write what is decided in the use
// cases, so the central guard lets members' writes through like personal
// settings.
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/announcements", h.List)
	r.Protected("POST /api/announcements", h.Create)
	r.Protected("GET /api/announcements/channels", h.Channels)
	r.Protected("GET /api/announcements/settings", h.Settings)
	r.Protected("PUT /api/announcements/settings", h.SaveSettings)
	r.Protected("PUT /api/announcements/{id}", h.Update)
	r.Protected("DELETE /api/announcements/{id}", h.Delete)
	r.Protected("POST /api/announcements/{id}/read", h.MarkRead)
}

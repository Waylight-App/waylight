package notifications

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"net/url"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/notifications/adapter/controller"
	"waylight/internal/modules/notifications/adapter/gateway"
	"waylight/internal/modules/notifications/domain"
	"waylight/internal/modules/notifications/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the notification rings, implements modules.Module and is the
// kernel.ObservationSink other modules report to.
type Module struct {
	modules.Base
	uc       usecase.Notifier
	email    *gateway.EmailPublisher
	handlers *controller.Handlers
	devices  *controller.AppDeviceHandlers
}

// Push configures delivery to the mobile app (#564): the relay's address
// (empty turns it off) and the app sessions the devices belong to.
type Push struct {
	RelayURL string
	Sessions kernel.AppSessions
}

// New builds the notifications module; cipher encrypts the ntfy token.
// publicURL (PUBLIC_URL, may be nil) is linked from Slack and Teams messages.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator, publicURL *url.URL, push Push) *Module {
	link := ""
	if publicURL != nil {
		link = publicURL.String()
	}
	email := gateway.NewEmailPublisher(clock.Now)
	var relay usecase.PushRelay
	if push.RelayURL != "" {
		relay = gateway.NewRelayClient(client, push.RelayURL, db, cipher, clock.Now)
	}
	devices := usecase.NewAppDevices(gateway.NewAppDeviceRepository(db, cipher), relay, push.Sessions, clock)
	uc := usecase.NewNotifier(usecase.Deps{
		Repo: gateway.NewRepository(db, cipher), Sender: gateway.NewNtfySender(client),
		Channels: gateway.NewChannelRepository(db, cipher), Maintenance: gateway.NewMaintenanceRepository(db),
		Pending: gateway.NewPendingRepository(db), Clock: clock, IDs: ids, Location: time.Local,
		Publishers: map[domain.ChannelKind]usecase.Publisher{
			domain.ChannelGotify:  gateway.NewGotifyPublisher(client),
			domain.ChannelWebhook: gateway.NewWebhookPublisher(client),
			domain.ChannelEmail:   email,
			domain.ChannelSlack:   gateway.NewSlackPublisher(client, link),
			domain.ChannelTeams:   gateway.NewTeamsPublisher(client, link),
			domain.ChannelApp:     devices,
		},
	})
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), devices: controller.NewAppDeviceHandlers(devices), email: email}
}

// SetInstanceName makes e-mail subjects start with the instance's name
// from the branding (#305) instead of the product's.
func (m *Module) SetInstanceName(name kernel.InstanceName) { m.email.Name = name }

// Alert implements kernel.AlertSink.
func (m *Module) Alert(ctx context.Context, a kernel.Alert) { m.uc.Alert(ctx, a) }

// Observe implements kernel.ObservationSink.
func (m *Module) Observe(ctx context.Context, o kernel.Observation) { m.uc.Observe(ctx, o) }

// CanSend implements kernel.Mailer (#296).
func (m *Module) CanSend(ctx context.Context) bool { return m.uc.CanSend(ctx) }

// Send implements kernel.Mailer (#296).
func (m *Module) Send(ctx context.Context, mail kernel.Mail) error { return m.uc.Send(ctx, mail) }

// ChatChannel is an enabled Slack or Teams channel (#332).
type ChatChannel = usecase.ChatChannel

// ChatChannels lists the enabled Slack and Teams channels, for modules that
// let an author post to one of them.
func (m *Module) ChatChannels(ctx context.Context) ([]ChatChannel, error) {
	return m.uc.ChatChannels(ctx)
}

// PostToChat posts a message to one Slack or Teams channel, outside the
// notification rules.
func (m *Module) PostToChat(ctx context.Context, channelID, title, body, link string) error {
	return m.uc.PostToChat(ctx, channelID, title, body, link)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "notifications" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/notifications/settings", m.handlers.Settings)
	r.Protected("PUT /api/notifications/settings", m.handlers.SaveSettings)
	r.Protected("POST /api/notifications/test", m.handlers.SendTest)
	r.Protected("GET /api/notifications/channels", m.handlers.Channels)
	r.Protected("POST /api/notifications/channels", m.handlers.SaveChannel)
	r.Protected("PUT /api/notifications/channels/{id}", m.handlers.SaveChannel)
	r.Protected("DELETE /api/notifications/channels/{id}", m.handlers.DeleteChannel)
	r.Protected("POST /api/notifications/channels/{id}/test", m.handlers.TestChannel)
	r.Protected("PUT /api/notifications/app-device", m.devices.Register)
	r.Protected("DELETE /api/notifications/app-device", m.devices.Unregister)
	r.Protected("GET /api/notifications/app-devices", m.devices.List)
	r.Protected("DELETE /api/notifications/app-devices/{session}", m.devices.Remove)
	r.Protected("GET /api/notifications/maintenance", m.handlers.Maintenance)
	r.Protected("POST /api/notifications/maintenance", m.handlers.StartMaintenance)
	r.Protected("DELETE /api/notifications/maintenance/{id}", m.handlers.EndMaintenance)
}

// digestInterval: how often deferred notifications are checked after quiet hours.
const digestInterval = time.Minute

// Jobs implements modules.Module: sends the quiet-hours digest.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "notifications.digest", Interval: digestInterval, Run: m.uc.FlushDigest}}
}

// UnderMaintenance reports whether an object is muted by a maintenance
// window (used by the services module to exclude it from uptime, #81).
func (m *Module) UnderMaintenance(ctx context.Context, kind, key, label string, at time.Time) bool {
	return m.uc.UnderMaintenance(ctx, kind, key, label, at)
}

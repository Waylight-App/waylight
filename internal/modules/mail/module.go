package mail

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/mail/adapter/controller"
	"trisse-home/internal/modules/mail/adapter/gateway"
	"trisse-home/internal/modules/mail/usecase"
	"trisse-home/internal/shared/imap"
	"trisse-home/internal/shared/kernel"
)

// Module wires the mail rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the module; dial is the guarded dialer of the integration
// clients. Nothing connects until an account is saved or a widget asks.
func New(db *sql.DB, cipher gateway.Cipher, dial imap.DialFunc, ids kernel.IDGenerator, clock kernel.Clock) *Module {
	uc := usecase.New(gateway.NewRepository(db, cipher), gateway.NewIMAP(dial), ids, clock)
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "mail" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"mail"} }

// RegisterRoutes implements modules.Module. Account management is under
// /settings, which guests cannot reach.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/mail/inbox", h.Inbox)
	r.Protected("GET /api/mail/settings/accounts", h.Accounts)
	r.Protected("POST /api/mail/settings/accounts", h.CreateAccount)
	r.Protected("PUT /api/mail/settings/accounts/{id}", h.UpdateAccount)
	r.Protected("DELETE /api/mail/settings/accounts/{id}", h.DeleteAccount)
}

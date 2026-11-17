package phonelist

import (
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/phonelist/adapter/controller"
	"waylight/internal/modules/phonelist/adapter/gateway"
	"waylight/internal/modules/phonelist/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the phone list rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.Phonelist
	handlers *controller.Handlers
}

// New builds the phone list module.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewPhonelist(usecase.Deps{Repo: gateway.NewRepository(db, cipher), Book: gateway.NewCardDAV(client), Clock: clock, IDs: ids})
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, uc: uc, handlers: controller.NewHandlers(uc)}
}

// Jobs implements modules.Module: the address book is read every hour.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "phonelist.sync", Interval: time.Hour, Run: m.uc.Sync}}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "phonelist" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind implements modules.Module (ADR 0018): admins maintain one list
// for everyone; members and guests read it.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"phonelist"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/phonelist/entries", m.handlers.List)
	r.Protected("POST /api/phonelist/entries", m.handlers.Add)
	r.Protected("PUT /api/phonelist/entries/{id}", m.handlers.Update)
	r.Protected("DELETE /api/phonelist/entries/{id}", m.handlers.Delete)
	// Under /settings, which guests and members cannot reach (it holds the
	// address book password).
	r.Protected("GET /api/phonelist/settings", m.handlers.Settings)
	r.Protected("PUT /api/phonelist/settings", m.handlers.SaveSettings)
}

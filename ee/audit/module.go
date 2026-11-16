// Package audit is the audit log of the ee/ features (#303): who changed
// instance data or sharing, when, on which object; never the content.
//
// Ring: module wiring. module.go connects the rings below and implements
// modules.Module, modules.Gated and modules.ChangeSink; no business logic.
package audit

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"time"

	"trisse-home/ee/audit/adapter/controller"
	"trisse-home/ee/audit/adapter/gateway"
	"trisse-home/ee/audit/domain"
	"trisse-home/ee/audit/usecase"
	"trisse-home/ee/licence"
	"trisse-home/internal/modules"
	"trisse-home/internal/shared/kernel"
)

// Module wires the rings.
type Module struct {
	modules.Base
	svc       *usecase.Service
	handlers  *controller.Handlers
	licensing licence.Licensing
	log       *slog.Logger
}

// New builds the module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, licensing licence.Licensing, log *slog.Logger) *Module {
	svc := usecase.New(usecase.Deps{Store: gateway.NewStore(db), Clock: clock, IDs: ids, Licensing: licensing})
	return &Module{svc: svc, handlers: controller.NewHandlers(svc), licensing: licensing, log: log}
}

// RecordChange implements modules.ChangeSink. A failure is logged, never
// passed on: the change itself has already happened.
func (m *Module) RecordChange(ctx context.Context, c modules.Change) {
	err := m.svc.Record(ctx, domain.Entry{AccountID: c.AccountID, Role: c.Role, Module: c.Module, Route: c.Route, ObjectID: c.ObjectID})
	if err != nil && m.log != nil {
		m.log.Error("audit log: could not record a change", "module", c.Module, "route", c.Route, "err", err)
	}
}

// Access implements modules.Gated.
func (m *Module) Access(ctx context.Context) modules.Access {
	return licence.Gate(ctx, m.licensing, licence.FeatureAuditLog)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "audit" }

// DataKind implements modules.Module.
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// Jobs implements modules.Module: entries older than the retention go daily.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "audit.prune", Interval: 24 * time.Hour, Run: m.svc.Prune}}
}

// RegisterRoutes implements modules.Module. There is no route to change or
// delete an entry; admins only (checked in the use cases).
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/audit", m.handlers.List)
	r.Protected("GET /api/audit/export", m.handlers.Export)
	r.Protected("GET /api/audit/settings", m.handlers.Settings)
	r.Protected("PUT /api/audit/settings", m.handlers.SaveSettings)
}

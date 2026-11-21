package wireguard

import (
	"context"
	"database/sql"
	"io/fs"

	"waylight/internal/modules"
	"waylight/internal/modules/wireguard/adapter/controller"
	"waylight/internal/modules/wireguard/adapter/gateway"
	"waylight/internal/modules/wireguard/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the WireGuard rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	uc       usecase.WireGuard
	handlers *controller.Handlers
}

// New builds the WireGuard module; dir is the export directory.
func New(db *sql.DB, dir string, clock kernel.Clock) *Module {
	uc := usecase.NewWireGuard(gateway.NewFileSource(dir), gateway.NewNameStore(db), clock)
	return &Module{uc: uc, Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "wireguard" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"wireguard"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/wireguard/peers", m.handlers.List)
	r.Protected("PUT /api/wireguard/names", m.handlers.Rename)
}

// CompactState implements modules.CompactProvider (#641): connected peers,
// and whether the export is still current.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	v, err := m.uc.List(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	return usecase.CompactState(v), nil
}

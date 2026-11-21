package diskhealth

import (
	"context"
	"errors"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/diskhealth/adapter/controller"
	"waylight/internal/modules/diskhealth/adapter/gateway"
	"waylight/internal/modules/diskhealth/domain"
	"waylight/internal/modules/diskhealth/usecase"
	"waylight/internal/shared/kernel"
)

// evaluateInterval: how often disk states are reported to notifications.
const evaluateInterval = 5 * time.Minute

// Config configures the module.
type Config struct {
	// Dir is the export directory of scripts/smart-export.sh.
	Dir string
	// StaleAfter marks older exports as unknown (0 = default).
	StaleAfter time.Duration
}

// Module wires the disk health rings and implements modules.Module.
type Module struct {
	modules.Base
	uc       usecase.DiskHealth
	handlers *controller.Handlers
	clock    kernel.Clock
}

// New builds the disk health module; alerter receives state observations.
func New(cfg Config, alerter kernel.ObservationSink, clock kernel.Clock) *Module {
	uc := usecase.NewDiskHealth(gateway.NewFileSource(cfg.Dir), alerter, clock, cfg.StaleAfter)
	return &Module{uc: uc, handlers: controller.NewHandlers(uc), clock: clock}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "diskhealth" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"diskhealth"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/diskhealth", m.handlers.List)
}

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "diskhealth.evaluate", Interval: evaluateInterval, Run: m.uc.Evaluate}}
}

// CompactState implements modules.CompactProvider (#594): the health of
// all drives in one line.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	disks, err := m.uc.Disks(ctx)
	// Without an export yet the tile says so instead of failing.
	if err != nil && !errors.Is(err, domain.ErrNoExport) {
		return kernel.CompactState{}, err
	}
	return usecase.CompactState(disks, m.clock.Now()), nil
}

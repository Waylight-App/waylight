package documents

import (
	"context"
	"database/sql"
	"io/fs"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/documents/adapter/controller"
	"waylight/internal/modules/documents/adapter/gateway"
	"waylight/internal/modules/documents/domain"
	"waylight/internal/modules/documents/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the document rings and implements modules.Module (#688).
type Module struct {
	seedDemo func(ctx context.Context) error
	modules.Base
	handlers *controller.Handlers
	uc       usecase.Documents
	clock    kernel.Clock
}

// New builds the documents module; loc decides what "today" is.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, loc *time.Location) *Module {
	deps := usecase.Deps{Repo: gateway.NewRepository(db), Clock: clock, IDs: ids, Location: loc}
	uc := usecase.NewDocuments(deps)
	return &Module{
		seedDemo: func(ctx context.Context) error { return usecase.SeedDemo(ctx, deps) },
		handlers: controller.NewHandlers(uc),
		uc:       uc,
		clock:    clock,
	}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "documents" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"documents"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/documents", m.handlers.List)
	r.Protected("POST /api/documents", m.handlers.Create)
	r.Protected("PUT /api/documents/{id}", m.handlers.Update)
	r.Protected("DELETE /api/documents/{id}", m.handlers.Delete)
	r.Protected("POST /api/documents/{id}/file", m.handlers.UploadFile)
	r.Protected("GET /api/documents/{id}/file", m.handlers.File)
}

// CompactState implements modules.CompactProvider (#597): the warranty that
// ends first; attention when some end within 30 days.
func (m *Module) CompactState(ctx context.Context) (kernel.CompactState, error) {
	list, err := m.uc.List(ctx)
	if err != nil {
		return kernel.CompactState{}, err
	}
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: m.clock.Now()}
	soon := 0
	for _, d := range list {
		if d.Status == domain.EndingSoon {
			soon++
		}
	}
	// The list puts covered warranties first, the soonest end on top.
	if len(list) == 0 || (list[0].Status != domain.Running && list[0].Status != domain.EndingSoon) {
		state.Title = kernel.NewText("No warranties running")
		state.Empty = true
		return state, nil
	}
	next := list[0]
	state.Title = kernel.NewTextf("{name}: warranty ends {date}", "name", next.Title, "date", domain.Day(next.WarrantyEnd))
	if soon > 0 {
		state.Level = kernel.CompactAttention
		state.Count = kernel.IntPtr(soon)
	}
	return state, nil
}

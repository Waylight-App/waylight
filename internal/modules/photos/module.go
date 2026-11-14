package photos

import (
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/photos/adapter/controller"
	"trisse-home/internal/modules/photos/adapter/gateway"
	"trisse-home/internal/modules/photos/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the photo frame rings and implements modules.Module.
//
// It is deliberately not Transferable: data exports are JSON and would
// carry hundreds of megabytes of pictures. Backups (a copy of the
// database) contain them.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the photo frame module.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewPhotos(usecase.Deps{Repo: gateway.NewRepository(db), Processor: gateway.Processor{}, Clock: clock, IDs: ids})
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "photos" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"photo-frame"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/photos", h.List)
	r.Protected("GET /api/photos/albums", h.Albums)
	r.Protected("POST /api/photos", h.Upload)
	r.Protected("PUT /api/photos/order", h.Reorder)
	r.Protected("DELETE /api/photos/{id}", h.Delete)
	r.Protected("GET /api/photos/{id}/image", h.Image)
}

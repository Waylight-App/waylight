package arr

import (
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/arr/adapter/controller"
	"waylight/internal/modules/arr/adapter/gateway"
	"waylight/internal/modules/arr/domain"
	"waylight/internal/modules/arr/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the *arr rings and implements modules.Module.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the *arr module (#86); cipher encrypts the API keys.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.NewArr(usecase.Deps{
		Repo: gateway.NewInstanceRepository(db, cipher),
		Providers: map[domain.App]usecase.Provider{
			domain.AppSonarr: gateway.NewSonarr(client), domain.AppRadarr: gateway.NewRadarr(client), domain.AppProwlarr: gateway.NewProwlarr(client),
		},
		Clock: clock, IDs: ids,
	})
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "arr" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"arr"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/arr/instances", m.handlers.Instances)
	r.Protected("POST /api/arr/instances", m.handlers.Add)
	r.Protected("PUT /api/arr/instances/{id}", m.handlers.Update)
	r.Protected("DELETE /api/arr/instances/{id}", m.handlers.Delete)
	r.Protected("GET /api/arr/overview", m.handlers.Overview)
}

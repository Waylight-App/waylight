package wiki

import (
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/wiki/adapter/controller"
	"trisse-home/internal/modules/wiki/adapter/gateway"
	"trisse-home/internal/modules/wiki/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the wiki search rings and implements modules.Module. It has
// no widget: its results appear in the smart search.
type Module struct {
	modules.Base
	handlers *controller.Handlers
}

// New builds the wiki search module. client is the integration client
// (wikis often run in the home network); cipher encrypts the API tokens.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator) *Module {
	uc := usecase.New(usecase.Deps{Repo: gateway.NewRepository(db, cipher), Searchers: gateway.Searchers(client, clock.Now), IDs: ids, Clock: clock})
	return &Module{handlers: controller.NewHandlers(uc)}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "wiki" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module. Connections carry tokens, so
// they are under the owner's session like other integration settings.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/wiki/search", h.Search)
	r.Protected("GET /api/wiki/wikis", h.List)
	r.Protected("POST /api/wiki/wikis", h.Create)
	r.Protected("PUT /api/wiki/wikis/{id}", h.Update)
	r.Protected("DELETE /api/wiki/wikis/{id}", h.Delete)
	r.Protected("POST /api/wiki/wikis/{id}/test", h.Test)
}

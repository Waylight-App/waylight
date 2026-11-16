package wiki

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/wiki/adapter/controller"
	"trisse-home/internal/modules/wiki/adapter/gateway"
	"trisse-home/internal/modules/wiki/domain"
	"trisse-home/internal/modules/wiki/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the wiki search rings and implements modules.Module. It has
// no widget: its results appear in the smart search.
type Module struct {
	modules.Base
	handlers *controller.Handlers
	uc       usecase.Wikis
}

// Page is a wiki page found by Find.
type Page = domain.Page

// Find searches the connected wikis and the member sources for the account
// in ctx, for searches outside the app such as a chat command (#333).
func (m *Module) Find(ctx context.Context, query string) ([]Page, error) {
	res, err := m.uc.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	return res.Pages, nil
}

// New builds the wiki search module. client is the integration client
// (wikis often run in the home network); cipher encrypts the API tokens.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, clock kernel.Clock, ids kernel.IDGenerator, members *MemberSources) *Module {
	uc := usecase.New(usecase.Deps{Repo: gateway.NewRepository(db, cipher), Searchers: gateway.Searchers(client, clock.Now), IDs: ids, Clock: clock,
		Members: members})
	return &Module{handlers: controller.NewHandlers(uc), uc: uc}
}

// Name implements modules.Module.
func (m *Module) Name() string { return "wiki" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.InstanceData }

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

// MemberSources holds the wikis searched on behalf of the signed-in person;
// see usecase.MemberSources.
type MemberSources = usecase.MemberSources

// MemberSource searches a wiki for the signed-in person; see
// usecase.MemberSource.
type MemberSource = usecase.MemberSource

// ErrNotConnected skips a member source quietly.
var ErrNotConnected = usecase.ErrNotConnected

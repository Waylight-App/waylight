package shopping

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"

	"waylight/internal/modules"
	"waylight/internal/modules/shopping/adapter/controller"
	"waylight/internal/modules/shopping/adapter/gateway"
	"waylight/internal/modules/shopping/usecase"
	"waylight/internal/shared/kernel"
)

// Module wires the shopping list rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	lists    *controller.ListHandlers
	repo     *gateway.ListRepository
	// listsUC fills the demo's list (SeedDemo).
	listsUC *usecase.Lists
}

// New builds the shopping list module; cipher encrypts the app password.
// host (from PUBLIC_URL) makes the UIDs of exported items unique.
func New(db *sql.DB, cipher gateway.Cipher, client *http.Client, ids kernel.IDGenerator, clock kernel.Clock, host string) *Module {
	settings, store := gateway.NewSettingsRepository(db, cipher), gateway.NewCalDAVStore(client)
	uc := usecase.NewShopping(settings, store, ids)
	repo := gateway.NewListRepository(db)
	if host == "" {
		host = "waylight.local"
	}
	lists := usecase.NewLists(repo, settings, store, ids, clock, host)
	return &Module{
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(uc),
		lists:        controller.NewListHandlers(lists),
		listsUC:      lists,
		repo:         repo,
	}
}

// ForgetAccount removes a deleted account's lists and their items (#471).
func (m *Module) ForgetAccount(ctx context.Context, accountID string) error {
	return m.repo.DeleteOwnedBy(ctx, accountID)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "shopping" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"shopping"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/shopping/settings", h.Settings)
	r.Protected("PUT /api/shopping/settings", h.SaveSettings)
	r.Protected("GET /api/shopping/items", h.List)
	r.Protected("POST /api/shopping/items", h.Add)
	r.Protected("PUT /api/shopping/items/{id}", h.SetDone)
	r.Protected("DELETE /api/shopping/items/{id}", h.Delete)
	r.Protected("POST /api/shopping/clear-completed", h.ClearCompleted)
	// Lists kept on Waylight (#471).
	l := m.lists
	r.Protected("GET /api/shopping/lists", l.All)
	r.Protected("POST /api/shopping/lists", l.Create)
	r.Protected("PUT /api/shopping/lists/{id}", l.Update)
	r.Protected("DELETE /api/shopping/lists/{id}", l.Delete)
	r.Protected("PUT /api/shopping/current", l.SetCurrent)
	r.Protected("GET /api/shopping/lists/{id}/items", l.Items)
	r.Protected("POST /api/shopping/lists/{id}/items", l.Add)
	r.Protected("PUT /api/shopping/lists/{id}/items/{itemId}", l.SetDone)
	r.Protected("DELETE /api/shopping/lists/{id}/items/{itemId}", l.DeleteItem)
	r.Protected("POST /api/shopping/lists/{id}/clear-completed", l.ClearCompleted)
	r.Protected("POST /api/shopping/lists/{id}/import", l.Import)
	r.Protected("GET /api/shopping/lists/{id}/export", l.Export)
}

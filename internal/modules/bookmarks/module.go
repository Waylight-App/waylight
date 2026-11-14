package bookmarks

import (
	"context"
	"database/sql"
	"io/fs"
	"net/http"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/bookmarks/adapter/configfile"
	"trisse-home/internal/modules/bookmarks/adapter/controller"
	"trisse-home/internal/modules/bookmarks/adapter/gateway"
	"trisse-home/internal/modules/bookmarks/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the bookmarks rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	seed     usecase.SeedSampleData
	check    usecase.CheckLinks
	config   usecase.Config
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

// linkCheckInterval is how often the link check job runs; each run checks
// at most 25 links that were last checked a week ago or never (#72).
const linkCheckInterval = time.Hour

// Jobs implements modules.Module.
func (m *Module) Jobs() []modules.Job {
	return []modules.Job{{Name: "bookmarks.links", Interval: linkCheckInterval, Run: func(ctx context.Context) error {
		_, err := m.check.Execute(ctx)
		return err
	}}}
}

// New builds the bookmarks module. client is used for the opt-in title,
// favicon and link checks and should carry the internal address guard.
func New(db *sql.DB, clock kernel.Clock, ids kernel.IDGenerator, client *http.Client) *Module {
	repo := gateway.NewRepository(db)
	d := usecase.Deps{Repo: repo, Clock: clock, IDs: ids, Settings: repo, Links: repo, Icons: repo}
	config := usecase.NewConfig(repo, clock, ids)
	if client != nil {
		web := gateway.NewWeb(client)
		d.Metadata, d.Checker = web, web
	}
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables}, seed: usecase.NewSeedSampleData(d), config: config,
		check: usecase.NewCheckLinks(d), handlers: controller.NewHandlers(controller.UseCases{
			GetSettings: usecase.NewGetSettings(d), SaveSettings: usecase.NewSaveSettings(d), Enrich: usecase.NewEnrichBookmark(d),
			CheckLink: usecase.NewCheckLink(d), Icon: usecase.NewGetIcon(d),
			List: usecase.NewListBookmarks(d), Create: usecase.NewCreateBookmark(d), Update: usecase.NewUpdateBookmark(d),
			Delete: usecase.NewDeleteBookmark(d), CreateFolder: usecase.NewCreateFolder(d),
			RenameFolder: usecase.NewRenameFolder(d), DeleteFolder: usecase.NewDeleteFolder(d),
			Import: usecase.NewImportBookmarks(d),
		})}
}

// ConfigSection implements modules.Configurable.
func (m *Module) ConfigSection() string { return "bookmarks" }

// ExportConfig implements modules.Configurable.
func (m *Module) ExportConfig(ctx context.Context) (any, error) {
	s, err := m.config.Export(ctx)
	if err != nil {
		return nil, err
	}
	return configfile.FromSection(s), nil
}

// NewConfig implements modules.Configurable.
func (m *Module) NewConfig() any { return &configfile.Section{} }

// ApplyConfig implements modules.Configurable.
func (m *Module) ApplyConfig(ctx context.Context, section any) error {
	s, ok := section.(*configfile.Section)
	if !ok {
		return modules.ErrConfigType
	}
	cs, err := s.Section()
	if err != nil {
		return err
	}
	return m.config.Apply(ctx, cs)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "bookmarks" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// WidgetKinds implements modules.Module.
func (m *Module) WidgetKinds() []string { return []string{"bookmarks"} }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/bookmarks", h.List)
	r.Protected("GET /api/bookmarks/settings", h.GetSettings)
	r.Protected("PUT /api/bookmarks/settings", h.SaveSettings)
	r.Protected("POST /api/bookmarks/{id}/enrich", h.Enrich)
	r.Protected("POST /api/bookmarks/{id}/check", h.CheckLink)
	r.Protected("GET /api/bookmarks/{id}/icon", h.Icon)
	r.Protected("POST /api/bookmarks", h.Create)
	r.Protected("PUT /api/bookmarks/{id}", h.Update)
	r.Protected("DELETE /api/bookmarks/{id}", h.Delete)
	r.Protected("POST /api/bookmarks/folders", h.CreateFolder)
	r.Protected("PATCH /api/bookmarks/folders/{id}", h.RenameFolder)
	r.Protected("DELETE /api/bookmarks/folders/{id}", h.DeleteFolder)
	r.Protected("POST /api/bookmarks/import", h.Import)
}

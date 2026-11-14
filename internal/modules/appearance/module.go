package appearance

import (
	"context"
	"database/sql"
	"io/fs"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/appearance/adapter/configfile"
	"trisse-home/internal/modules/appearance/adapter/controller"
	"trisse-home/internal/modules/appearance/adapter/gateway"
	"trisse-home/internal/modules/appearance/usecase"
)

// Module wires the appearance rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	themes   *controller.ThemeHandlers
	repo     usecase.Repository
	get      usecase.GetAppearance
	update   usecase.UpdateAppearance
}

// SetProfile stores the start-page profile chosen at the first setup
// (#281; called from the composition root).
func (m *Module) SetProfile(ctx context.Context, profile string) error {
	a, err := m.repo.Load(ctx)
	if err != nil {
		return err
	}
	if err := a.SetProfile(profile); err != nil {
		return err
	}
	return m.repo.Save(ctx, a)
}

// New builds the appearance module.
func New(db *sql.DB) *Module {
	repo := gateway.NewRepository(db)
	themes := usecase.NewThemes(repo, gateway.NewThemeStore(db), gateway.Codec{}, gateway.BuiltinThemes())
	get, update := usecase.NewGetAppearance(repo), usecase.NewUpdateAppearance(repo)
	return &Module{
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(get, update),
		get:          get,
		update:       update,
		themes:       controller.NewThemeHandlers(themes),
		repo:         repo,
	}
}

// ConfigSection implements modules.Configurable.
func (m *Module) ConfigSection() string { return "appearance" }

// ExportConfig implements modules.Configurable.
func (m *Module) ExportConfig(ctx context.Context) (any, error) {
	r, err := m.get.Execute(ctx)
	if err != nil {
		return nil, err
	}
	return configfile.FromResponse(r), nil
}

// ConfigRoutes implements modules.Configurable.
func (m *Module) ConfigRoutes() []string {
	return []string{
		"PUT /api/appearance",
		"POST /api/appearance/themes",
		"POST /api/appearance/themes/restore",
		"DELETE /api/appearance/themes/{id}",
		"POST /api/appearance/themes/{id}/apply",
	}
}

// NewConfig implements modules.Configurable.
func (m *Module) NewConfig() any { return &configfile.Section{} }

// ApplyConfig implements modules.Configurable.
func (m *Module) ApplyConfig(ctx context.Context, section any) error {
	s, ok := section.(*configfile.Section)
	if !ok {
		return modules.ErrConfigType
	}
	_, err := m.update.Execute(ctx, s.Request())
	return err
}

// Name implements modules.Module.
func (m *Module) Name() string { return "appearance" }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	r.Protected("GET /api/appearance", m.handlers.Get)
	r.Protected("PUT /api/appearance", m.handlers.Update)
	r.Protected("GET /api/appearance/themes", m.themes.Gallery)
	r.Protected("POST /api/appearance/themes", m.themes.Import)
	r.Protected("POST /api/appearance/themes/restore", m.themes.Restore)
	r.Protected("GET /api/appearance/themes/{id}", m.themes.Export)
	r.Protected("DELETE /api/appearance/themes/{id}", m.themes.Delete)
	r.Protected("POST /api/appearance/themes/{id}/apply", m.themes.Apply)
}

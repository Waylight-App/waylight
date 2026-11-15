package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/fs"
	"time"

	"trisse-home/internal/modules"
	"trisse-home/internal/modules/dashboard/adapter/configfile"
	"trisse-home/internal/modules/dashboard/adapter/controller"
	"trisse-home/internal/modules/dashboard/adapter/gateway"
	"trisse-home/internal/modules/dashboard/usecase"
	"trisse-home/internal/shared/kernel"
)

// Module wires the dashboard rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	seed     usecase.SeedSampleData
	everyday usecase.SeedSampleData
	scope    usecase.TabScope
	config   usecase.Config
}

// SampleWeather is the location of the example weather widget.
type SampleWeather struct {
	Label    string
	Lat, Lon float64
}

// New builds the dashboard module. widgetKinds lists the widget kinds of all
// enabled modules.
func New(db *sql.DB, ids kernel.IDGenerator, clock kernel.Clock, widgetKinds []string, weather SampleWeather) *Module {
	repo := gateway.NewRepository(db)
	deps := usecase.Deps{Tabs: repo, Widgets: repo, Templates: repo, Kinds: gateway.NewKindCatalog(widgetKinds), IDs: ids,
		Clock: clock, Location: time.Local}
	weatherConfig, _ := json.Marshal(map[string]any{"lat": weather.Lat, "lon": weather.Lon, "label": weather.Label, "days": 4}) // plain map, cannot fail
	return &Module{Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers: controller.NewHandlers(usecase.NewGetDashboard(deps), usecase.NewManageTabs(deps), usecase.NewManageWidgets(deps),
			usecase.NewManageTemplates(deps)),
		seed:     usecase.NewSeedSampleData(deps, usecase.SampleLayout(string(weatherConfig))),
		everyday: usecase.NewSeedSampleData(deps, usecase.EverydayLayout(string(weatherConfig))),
		scope:    usecase.NewTabScope(deps),
		config:   usecase.NewConfig(repo, deps.Kinds, ids),
	}
}

// ConfigSection implements modules.Configurable.
func (m *Module) ConfigSection() string { return "dashboard" }

// ExportConfig implements modules.Configurable.
func (m *Module) ExportConfig(ctx context.Context) (any, error) {
	tabs, err := m.config.Export(ctx)
	if err != nil {
		return nil, err
	}
	return configfile.FromTabs(tabs)
}

// ConfigRoutes implements modules.Configurable. Saving and deleting templates stays available; creating a tab from one does not.
func (m *Module) ConfigRoutes() []string {
	return []string{
		"POST /api/dashboard/tabs",
		"PUT /api/dashboard/tabs/order",
		"PATCH /api/dashboard/tabs/{id}",
		"DELETE /api/dashboard/tabs/{id}",
		"POST /api/dashboard/tabs/{id}/widgets",
		"PUT /api/dashboard/tabs/{id}/layout",
		"PATCH /api/dashboard/widgets/{id}",
		"DELETE /api/dashboard/widgets/{id}",
		"PUT /api/dashboard/tabs/{id}/background",
		"PUT /api/dashboard/tabs/{id}/rules",
		"PUT /api/dashboard/tabs/{id}/mobile-layout",
		"DELETE /api/dashboard/tabs/{id}/mobile-layout",
		"POST /api/dashboard/templates/{id}/tabs",
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
	tabs, err := s.ToTabs()
	if err != nil {
		return err
	}
	return m.config.Apply(ctx, tabs)
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

// SeedForProfile seeds the layout of a start-page profile (#281):
// "everyday" leaves out the Server tab.
func (m *Module) SeedForProfile(ctx context.Context, profile string) error {
	if profile == "everyday" {
		return m.everyday.Execute(ctx)
	}
	return m.seed.Execute(ctx)
}

// Name implements modules.Module.
func (m *Module) Name() string { return "dashboard" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalContent }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/dashboard", h.Get)
	r.Protected("POST /api/dashboard/tabs", h.CreateTab)
	r.Protected("PUT /api/dashboard/tabs/order", h.ReorderTabs)
	r.Protected("PATCH /api/dashboard/tabs/{id}", h.RenameTab)
	r.Protected("DELETE /api/dashboard/tabs/{id}", h.DeleteTab)
	r.Protected("POST /api/dashboard/tabs/{id}/widgets", h.AddWidget)
	r.Protected("PUT /api/dashboard/tabs/{id}/layout", h.SaveLayout)
	r.Protected("PATCH /api/dashboard/widgets/{id}", h.ConfigureWidget)
	r.Protected("DELETE /api/dashboard/widgets/{id}", h.DeleteWidget)
	r.Protected("PUT /api/dashboard/tabs/{id}/background", h.SetBackground)
	r.Protected("PUT /api/dashboard/tabs/{id}/rules", h.SetRules)
	r.Protected("PUT /api/dashboard/tabs/{id}/mobile-layout", h.SaveMobileLayout)
	r.Protected("DELETE /api/dashboard/tabs/{id}/mobile-layout", h.ResetMobileLayout)
	r.Protected("GET /api/dashboard/templates", h.ListTemplates)
	r.Protected("POST /api/dashboard/templates", h.SaveTemplate)
	r.Protected("DELETE /api/dashboard/templates/{id}", h.DeleteTemplate)
	r.Protected("POST /api/dashboard/templates/{id}/tabs", h.CreateTabFromTemplate)
}

// WidgetKindsOnTabs returns the widget kinds placed on the given tabs (used
// to scope guest accounts).
func (m *Module) WidgetKindsOnTabs(ctx context.Context, tabIDs []string) ([]string, error) {
	return m.scope.WidgetKindsOnTabs(ctx, tabIDs)
}

// TabIDs returns all tab ids (used for the display of an appliance, #140).
func (m *Module) TabIDs(ctx context.Context) ([]string, error) { return m.scope.TabIDs(ctx) }

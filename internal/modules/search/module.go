package search

import (
	"context"
	"database/sql"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/search/adapter/configfile"
	"waylight/internal/modules/search/adapter/controller"
	"waylight/internal/modules/search/adapter/gateway"
	"waylight/internal/modules/search/usecase"
	"waylight/internal/shared/answers"
	"waylight/internal/shared/kernel"
)

// Module wires the search rings and implements modules.Module.
type Module struct {
	modules.Base
	modules.Transferable
	handlers *controller.Handlers
	repo     *gateway.Repository
	seed     usecase.SeedSampleData
	config   usecase.Config
	answers  *answerEngine
}

// answerEngine answers the smart search over the sources the account may use.
type answerEngine struct {
	sources []answers.Source
	mayUse  func(ctx context.Context, module string) bool
	clock   kernel.Clock
	log     *slog.Logger
}

// Answer implements controller.AnswerEngine. Sources of modules hidden from
// the account by the admin's policy do not answer.
func (e *answerEngine) Answer(r *http.Request, query string, lang answers.Lang) []answers.Answer {
	ctx := r.Context()
	allowed := make([]answers.Source, 0, len(e.sources))
	for _, s := range e.sources {
		if e.mayUse == nil || e.mayUse(ctx, s.ID()) {
			allowed = append(allowed, s)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, answerTimeout)
	defer cancel()
	return answers.NewEngine(e.log, allowed...).Answer(ctx, query, lang, e.clock.Now())
}

// answerTimeout bounds one search; a slow source answers "not available".
const answerTimeout = 5 * time.Second

// SetAnswerSources installs the sources of the smart search (ADR 0026):
// those of the enabled modules, filtered per request by mayUse.
func (m *Module) SetAnswerSources(log *slog.Logger, clock kernel.Clock, mayUse func(ctx context.Context, module string) bool, sources []answers.Source) {
	m.answers = &answerEngine{sources: sources, mayUse: mayUse, clock: clock, log: log}
}

// New builds the search module.
func New(db *sql.DB) *Module {
	repo := gateway.NewRepository(db)
	return &Module{
		Transferable: modules.Transferable{DB: db, Tables: gateway.TransferTables},
		handlers:     controller.NewHandlers(usecase.NewSearcher(repo)),
		repo:         repo,
		seed:         usecase.NewSeedSampleData(repo),
		config:       usecase.NewConfig(repo),
	}
}

// ForgetAccount removes the search settings, shortcuts and own keywords of
// a deleted account (#478).
func (m *Module) ForgetAccount(ctx context.Context, accountID string) error {
	return m.repo.ForgetAccount(ctx, accountID)
}

// ConfigSection implements modules.Configurable.
func (m *Module) ConfigSection() string { return "search" }

// ExportConfig implements modules.Configurable.
func (m *Module) ExportConfig(ctx context.Context) (any, error) {
	s, err := m.config.Export(ctx)
	if err != nil {
		return nil, err
	}
	return configfile.FromSection(s), nil
}

// ConfigRoutes implements modules.Configurable.
func (m *Module) ConfigRoutes() []string {
	return []string{
		"PUT /api/search/settings",
		"PUT /api/search/bar",
		"POST /api/search/shortcuts",
		"DELETE /api/search/shortcuts/{keyword}",
		"POST /api/search/keywords",
		"DELETE /api/search/keywords/{keyword}",
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
	return m.config.Apply(ctx, s.Section())
}

// SeedSampleData implements modules.Seeder.
func (m *Module) SeedSampleData(ctx context.Context) error { return m.seed.Execute(ctx) }

// Name implements modules.Module.
func (m *Module) Name() string { return "search" }

// DataKind declares who owns the module's data (ADR 0018).
func (m *Module) DataKind() modules.DataKind { return modules.PersonalSettings }

// Migrations implements modules.Module.
func (m *Module) Migrations() fs.FS { return gateway.Migrations() }

// RegisterRoutes implements modules.Module.
func (m *Module) RegisterRoutes(r modules.Router) {
	h := m.handlers
	r.Protected("GET /api/search", h.Search)
	r.Public("GET /api/search/opensearch.xml", h.OpenSearch)
	r.Protected("GET /api/search/settings", h.Settings)
	a := controller.NewAnswerHandlers(func() controller.AnswerEngine {
		if m.answers == nil {
			return nil
		}
		return m.answers
	})
	r.Protected("GET /api/search/answers", a.Answers)
	r.Protected("PUT /api/search/settings", h.SetDefaultProvider)
	r.Protected("PUT /api/search/bar", h.SetBarOptions)
	r.Protected("POST /api/search/shortcuts", h.AddShortcut)
	r.Protected("DELETE /api/search/shortcuts/{keyword}", h.DeleteShortcut)
	r.Protected("POST /api/search/keywords", h.AddKeyword)
	r.Protected("DELETE /api/search/keywords/{keyword}", h.DeleteKeyword)
}

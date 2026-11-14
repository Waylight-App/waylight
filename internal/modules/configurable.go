package modules

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// Configurable is implemented by modules whose settings can be written down
// as code in trisse-home.yaml (#88). Unlike Portable, which dumps tables for
// backups, a config section is semantic: names instead of IDs, order given
// by list order, no secrets and no runtime state.
//
// The values are plain structs with yaml tags. The transfer module encodes
// and decodes them, so modules never depend on the file format.
type Configurable interface {
	// ConfigSection is the top-level key of the module in the file.
	ConfigSection() string
	// ExportConfig returns the current section.
	ExportConfig(ctx context.Context) (any, error)
	// NewConfig returns a pointer to an empty section to decode into.
	NewConfig() any
	// ApplyConfig validates the whole section (as returned by NewConfig
	// and filled by the decoder) and then replaces the module's settings
	// with it; on a validation error nothing is changed.
	ApplyConfig(ctx context.Context, section any) error
	// ConfigRoutes lists the routes ("METHOD /pattern", exactly as
	// registered) that change data owned by the section. They are refused
	// while CONFIG_FILE manages the section; runtime actions such as
	// Wake-on-LAN or a link check are not listed and keep working.
	ConfigRoutes() []string
}

// Configurables returns all registered modules that take part in config
// as code, keyed by their section name.
func (r *Registry) Configurables() map[string]Configurable {
	out := map[string]Configurable{}
	for _, m := range r.modules {
		if c, ok := m.(Configurable); ok {
			out[c.ConfigSection()] = c
		}
	}
	return out
}

// ErrConfigType means ApplyConfig got a value it did not create with
// NewConfig; that is a wiring bug, not a user error.
var ErrConfigType = errors.New("config section of the wrong type")

// ErrManaged answers a change to settings that CONFIG_FILE manages (#88).
var ErrManaged = kernel.NewPublicError(kernel.ErrConflict, "These settings are managed by the configuration file (CONFIG_FILE) and are read-only here.")

// ManagedRouter registers routes like the wrapped router, but refuses the
// listed config routes with 409, so a managed section cannot drift from
// the file between restarts.
func ManagedRouter(r Router, routes []string) Router { return managedRouter{r, routes} }

type managedRouter struct {
	Router
	routes []string
}

func (m managedRouter) Protected(pattern string, h http.HandlerFunc) {
	if slices.Contains(m.routes, pattern) {
		h = func(w http.ResponseWriter, r *http.Request) { httpapi.WriteDomainError(w, r, ErrManaged) }
	}
	m.Router.Protected(pattern, h)
}

package modules

import (
	"context"
	"errors"
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

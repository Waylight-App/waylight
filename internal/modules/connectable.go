package modules

import "context"

// CredentialField is one field of a service's credential, e.g. "apiKey".
// The name is the key in the integrations store (#610) and in the module's
// settings page; clients label it from their own texts.
type CredentialField struct {
	Name string
	// Secret fields are never answered by the API; other fields (a user
	// name, a token ID, a certificate fingerprint) are shown again.
	Secret bool
	// Optional fields may stay empty, e.g. the FRITZ!Box login.
	Optional bool
}

// ConnectionSpec says, as data, what a named connection to a service holds
// and how it is checked (#610).
type ConnectionSpec struct {
	// Service is the stable identifier stored with a connection, e.g.
	// "immich" or "sonarr"; a module may own several services.
	Service string
	// Title is the product name in its own spelling, e.g. "Home Assistant".
	Title  string
	Fields []CredentialField
	// FixedURL is the address of services that have exactly one, e.g. the
	// Tankerkönig API; connections of such a service cannot be moved.
	FixedURL string
	// Check tries the address and credential with the request the module's
	// settings page makes before saving. Its errors are public errors from
	// kernel.SetupFailure, so clients can show them as they are.
	Check func(ctx context.Context, url string, credential map[string]string) error
}

// Connectable is implemented by modules whose services can be reached
// through a named connection of the integrations module.
type Connectable interface {
	ConnectionSpecs() []ConnectionSpec
}

// ConnectionSpecsOf collects the specs of the given modules in their order.
func ConnectionSpecsOf(all []Module) []ConnectionSpec {
	var out []ConnectionSpec
	for _, m := range all {
		if c, ok := m.(Connectable); ok {
			out = append(out, c.ConnectionSpecs()...)
		}
	}
	return out
}

// SavedConnection is a connection a module keeps in its own settings from
// before the integrations store (#610). The store copies it once; the
// module's settings stay as they are.
type SavedConnection struct {
	// Source identifies the setting across restarts, e.g. "immich" or
	// "arr:<instance ID>"; a source is copied at most once.
	Source  string
	Service string
	// Name is the name of the copy; empty uses the product name.
	Name       string
	URL        string
	Credential map[string]string
}

// ConnectionSource is implemented by Connectable modules that still keep
// connection settings of their own. A module that is not configured
// returns none; an error skips the module until the next run.
type ConnectionSource interface {
	SavedConnections(ctx context.Context) ([]SavedConnection, error)
}

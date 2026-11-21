// Package configfile maps the connections to their section of waylight.yaml
// (#88, #610) and back. Credentials are never part of it.
//
// Ring: Interface Adapters.
package configfile

import "waylight/internal/modules/integrations/usecase"

// Connection is one connection by service and name.
type Connection struct {
	Service string `yaml:"service"`
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
}

// Section is the connections section, ordered by service and name.
type Section []Connection

// FromList maps the use case list.
func FromList(list []usecase.ConfigConnection) *Section {
	out := Section{}
	for _, c := range list {
		out = append(out, Connection(c))
	}
	return &out
}

// List maps back to the use case list.
func (s Section) List() []usecase.ConfigConnection {
	out := make([]usecase.ConfigConnection, 0, len(s))
	for _, c := range s {
		out = append(out, usecase.ConfigConnection(c))
	}
	return out
}

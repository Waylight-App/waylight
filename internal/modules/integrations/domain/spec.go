package domain

import (
	"maps"
	"slices"

	"waylight/internal/shared/kernel"
)

// Field is one credential field a service expects.
type Field struct {
	Name             string
	Secret, Optional bool
}

// Spec describes the connections of one service: which credential fields
// they hold and whether the service has a fixed address. The modules
// declare their specs; the store only checks connections against them.
type Spec struct {
	Service, Title string
	Fields         []Field
	FixedURL       string
}

// Errors of the specs.
var (
	ErrUnknownService = kernel.NewValidationError("service", "This service cannot be connected.")
	ErrUnknownField   = kernel.NewValidationError("credential", "The service does not expect this credential field.")
	ErrMissingField   = kernel.NewValidationError("credential", "Fill in every required credential field.")
)

// CheckCredential refuses fields the service does not know, and, when
// complete is set, a credential that misses a required field. Connections
// written by config as code have no credential yet, so complete is only
// set where an admin enters one.
func (s Spec) CheckCredential(c Credential, complete bool) error {
	for name := range c {
		if !slices.ContainsFunc(s.Fields, func(f Field) bool { return f.Name == name }) {
			return ErrUnknownField
		}
	}
	if complete {
		for _, f := range s.Fields {
			if !f.Optional && c[f.Name] == "" {
				return ErrMissingField
			}
		}
	}
	return nil
}

// Visible returns the fields that are not secret, e.g. the user name, so
// an admin can tell connections apart without seeing a password.
func (s Spec) Visible(c Credential) map[string]string {
	out := map[string]string{}
	for _, f := range s.Fields {
		if v := c[f.Name]; !f.Secret && v != "" {
			out[f.Name] = v
		}
	}
	return out
}

// Specs looks specs up by service.
type Specs map[string]Spec

// NewSpecs indexes specs by service; a later spec of the same service
// replaces an earlier one, which a test of the registry rules out.
func NewSpecs(list []Spec) Specs {
	out := Specs{}
	for _, s := range list {
		s.Fields = slices.Clone(s.Fields)
		out[s.Service] = s
	}
	return out
}

// Sorted lists the specs ordered by service.
func (s Specs) Sorted() []Spec {
	out := make([]Spec, 0, len(s))
	for _, k := range slices.Sorted(maps.Keys(s)) {
		out = append(out, s[k])
	}
	return out
}

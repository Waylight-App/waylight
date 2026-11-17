package usecase

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"waylight/internal/modules/transfer/domain"
	"waylight/internal/shared/kernel"
)

// ConfigSource is a module that takes part in config as code (#88); it is
// satisfied by modules.Configurable.
type ConfigSource interface {
	ConfigSection() string
	ExportConfig(ctx context.Context) (any, error)
	NewConfig() any
	ApplyConfig(ctx context.Context, section any) error
}

// NamedSection is one top-level section of the file.
type NamedSection struct {
	Name  string
	Value any
}

// ConfigCodec reads and writes waylight.yaml (output port).
type ConfigCodec interface {
	Encode(version int, sections []NamedSection) ([]byte, error)
	// Decode parses data, runs every string through expand and decodes
	// each section into the value returned by target; target returns
	// false for a section name nobody handles.
	Decode(data []byte, expand func(string) (string, error), target func(name string) (any, bool)) (version int, sections []NamedSection, err error)
}

// ApplyResult reports what applying a file did per section.
type ApplyResult struct {
	Applied []string
	// Failed maps section names to the reason; those sections are unchanged.
	Failed map[string]string
}

// Config exports and applies the config file (input port).
type Config interface {
	Export(ctx context.Context) ([]byte, error)
	// Apply applies a file uploaded by the owner; managed sections are
	// refused and reported as failed.
	Apply(ctx context.Context, data []byte) (ApplyResult, error)
	// ApplyManaged applies CONFIG_FILE at start-up. Every section must
	// succeed; the sections in the file then become managed.
	ApplyManaged(ctx context.Context, data []byte) ([]string, error)
	// Managed lists the sections CONFIG_FILE owns.
	Managed() []string
}

// errManagedSection is the reason reported for a section CONFIG_FILE owns.
const errManagedSection = "managed by CONFIG_FILE"

type configFile struct {
	sources func() map[string]ConfigSource
	codec   ConfigCodec
	env     func(string) (string, bool)
	audit   kernel.AuditSink
	managed *Managed
}

// NewConfig returns the Config use case. sources is called per request so
// modules registered later are included; env looks up ${NAME} references.
func NewConfig(sources func() map[string]ConfigSource, codec ConfigCodec, env func(string) (string, bool), audit kernel.AuditSink, managed *Managed) Config {
	return &configFile{sources: sources, codec: codec, env: env, audit: audit, managed: managed}
}

func (uc *configFile) Managed() []string { return uc.managed.List() }

func (uc *configFile) ApplyManaged(ctx context.Context, data []byte) ([]string, error) {
	res, err := uc.apply(ctx, data, false)
	if err != nil {
		return nil, err
	}
	if len(res.Failed) > 0 {
		var parts []string
		for _, name := range domain.SectionOrder {
			if reason, ok := res.Failed[name]; ok {
				parts = append(parts, name+": "+reason)
			}
		}
		return nil, fmt.Errorf("CONFIG_FILE: %s", strings.Join(parts, "; "))
	}
	uc.managed.Set(res.Applied)
	return res.Applied, nil
}

func (uc *configFile) Export(ctx context.Context) ([]byte, error) {
	src := uc.sources()
	var sections []NamedSection
	for _, name := range domain.SectionOrder {
		s, ok := src[name]
		if !ok {
			continue // module disabled
		}
		v, err := s.ExportConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("export %s: %w", name, err)
		}
		sections = append(sections, NamedSection{Name: name, Value: v})
	}
	return uc.codec.Encode(domain.ConfigVersion, sections)
}

// Apply decodes the whole file before changing anything, so a typo or a
// missing variable leaves every section as it was. Sections missing from
// the file are left alone; present ones are replaced as a whole.
func (uc *configFile) Apply(ctx context.Context, data []byte) (ApplyResult, error) {
	return uc.apply(ctx, data, true)
}

func (uc *configFile) apply(ctx context.Context, data []byte, respectManaged bool) (ApplyResult, error) {
	src := uc.sources()
	expand := func(s string) (string, error) { return domain.ExpandEnv(s, uc.env) }
	target := func(name string) (any, bool) {
		s, ok := src[name]
		if !ok || !slices.Contains(domain.SectionOrder, name) {
			return nil, false
		}
		return s.NewConfig(), true
	}
	version, sections, err := uc.codec.Decode(data, expand, target)
	if err != nil {
		return ApplyResult{}, err
	}
	if err := domain.CheckConfigVersion(version); err != nil {
		return ApplyResult{}, err
	}
	present := map[string]any{}
	for _, s := range sections {
		present[s.Name] = s.Value
	}
	res := ApplyResult{Applied: []string{}, Failed: map[string]string{}}
	for _, name := range domain.SectionOrder {
		v, ok := present[name]
		if !ok {
			continue
		}
		if respectManaged && uc.managed.Has(name) {
			res.Failed[name] = errManagedSection
			continue
		}
		if err := src[name].ApplyConfig(ctx, v); err != nil {
			res.Failed[name] = err.Error()
			continue
		}
		res.Applied = append(res.Applied, name)
	}
	if len(res.Applied) > 0 {
		uc.audit.Audit(ctx, kernel.AuditEvent{Type: "config_applied", Detail: strings.Join(res.Applied, ", ")})
	}
	return res, nil
}

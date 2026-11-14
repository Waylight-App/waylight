package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
)

type source struct {
	name    string
	applied any
	fail    error
}

func (s *source) ConfigSection() string                      { return s.name }
func (s *source) ExportConfig(context.Context) (any, error)  { return s.name + "-value", nil }
func (s *source) NewConfig() any                             { return new(string) }
func (s *source) ApplyConfig(_ context.Context, v any) error { s.applied = v; return s.fail }

type codec struct {
	encoded []NamedSection
	decoded []NamedSection
	version int
	err     error
}

func (c *codec) Encode(_ int, s []NamedSection) ([]byte, error) {
	c.encoded = s
	return []byte("ok"), nil
}
func (c *codec) Decode(_ []byte, _ func(string) (string, error), target func(string) (any, bool)) (int, []NamedSection, error) {
	for _, s := range c.decoded {
		if _, ok := target(s.Name); !ok {
			return 0, nil, kernel.NewValidationError("file", "unknown")
		}
	}
	return c.version, c.decoded, c.err
}

type audit struct{ events []kernel.AuditEvent }

func (a *audit) Audit(_ context.Context, e kernel.AuditEvent) { a.events = append(a.events, e) }

func TestConfigOrderAndPartialFailure(t *testing.T) {
	dash := &source{name: "dashboard"}
	svc := &source{name: "services", fail: kernel.NewValidationError("url", "bad")}
	app := &source{name: "appearance"}
	srcs := func() map[string]ConfigSource {
		return map[string]ConfigSource{"dashboard": dash, "services": svc, "appearance": app}
	}
	c := &codec{version: 1, decoded: []NamedSection{{Name: "dashboard", Value: "d"}, {Name: "services", Value: "s"}}}
	a := &audit{}
	uc := NewConfig(srcs, c, func(string) (string, bool) { return "", false }, a, &Managed{})

	if _, err := uc.Export(context.Background()); err != nil || len(c.encoded) != 3 || c.encoded[0].Name != "appearance" || c.encoded[2].Name != "dashboard" {
		t.Fatalf("export order: %+v %v", c.encoded, err)
	}
	res, err := uc.Apply(context.Background(), nil)
	if err != nil || len(res.Applied) != 1 || res.Applied[0] != "dashboard" || res.Failed["services"] == "" {
		t.Fatalf("%+v %v", res, err)
	}
	if app.applied != nil {
		t.Fatal("a section missing from the file must be left alone")
	}
	if len(a.events) != 1 || a.events[0].Detail != "dashboard" {
		t.Fatalf("audit: %+v", a.events)
	}
	c.version = 2
	if _, err := uc.Apply(context.Background(), nil); !errors.Is(err, kernel.ErrInvalid) || dash.applied != "d" {
		t.Fatalf("newer version: %v", err)
	}
}

func TestManagedSections(t *testing.T) {
	dash := &source{name: "dashboard"}
	svc := &source{name: "services"}
	srcs := func() map[string]ConfigSource { return map[string]ConfigSource{"dashboard": dash, "services": svc} }
	c := &codec{version: 1, decoded: []NamedSection{{Name: "dashboard", Value: "d"}}}
	m := &Managed{}
	uc := NewConfig(srcs, c, func(string) (string, bool) { return "", false }, &audit{}, m)
	got, err := uc.ApplyManaged(context.Background(), nil)
	if err != nil || len(got) != 1 || !m.Has("dashboard") || m.Has("services") {
		t.Fatalf("%v %v %v", got, err, m.List())
	}
	// An upload must not override the file.
	dash.applied = nil
	c.decoded = []NamedSection{{Name: "dashboard", Value: "x"}, {Name: "services", Value: "s"}}
	res, _ := uc.Apply(context.Background(), nil)
	if dash.applied != nil || res.Failed["dashboard"] == "" || res.Applied[0] != "services" {
		t.Fatalf("%+v", res)
	}
	// At start-up every section has to succeed.
	svc.fail = kernel.NewValidationError("url", "bad")
	if _, err := uc.ApplyManaged(context.Background(), nil); err == nil {
		t.Fatal("a failing section must stop the start")
	}
}

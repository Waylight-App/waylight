package gateway

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"trisse-home/internal/modules/transfer/domain"
	"trisse-home/internal/modules/transfer/usecase"
	"trisse-home/internal/shared/kernel"
)

type sample struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags,omitempty,flow"`
}

func target(name string) (any, bool) {
	if name != "things" {
		return nil, false
	}
	return &[]sample{}, true
}

func noEnv(s string) (string, error) { return s, nil }

func TestEncodeIsStableAndDecodesBack(t *testing.T) {
	in := []usecase.NamedSection{{Name: "things", Value: []sample{{Name: "a", Tags: []string{"x"}}, {Name: "true"}}}}
	out, err := YAMLCodec{}.Encode(1, in)
	if err != nil {
		t.Fatal(err)
	}
	want := "# yaml-language-server: $schema=" + SchemaURL + "\nversion: 1\nthings:\n  - name: a\n    tags: [x]\n  - name: \"true\"\n"
	if string(out) != want {
		t.Fatalf("got\n%s\nwant\n%s", out, want)
	}
	v, sections, err := YAMLCodec{}.Decode(out, noEnv, target)
	if err != nil || v != 1 || len(sections) != 1 {
		t.Fatalf("%v %v %v", v, sections, err)
	}
	if got := *sections[0].Value.(*[]sample); got[1].Name != "true" {
		t.Fatalf("a quoted value must stay a string: %+v", got)
	}
}

func TestDecodeExpandsOnlyStringValues(t *testing.T) {
	expand := func(s string) (string, error) { return strings.ReplaceAll(s, "${X}", "ok"), nil }
	_, sections, err := YAMLCodec{}.Decode([]byte("version: 1\nthings:\n  - name: ${X}\n"), expand, target)
	if err != nil || (*sections[0].Value.(*[]sample))[0].Name != "ok" {
		t.Fatalf("%v %v", sections, err)
	}
	failing := func(string) (string, error) { return "", kernel.NewValidationError("env", "not set") }
	if _, _, err := (YAMLCodec{}).Decode([]byte("version: 1\nthings:\n  - name: ${X}\n"), failing, target); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal(err)
	}
}

func TestDecodeRejects(t *testing.T) {
	for name, file := range map[string]string{
		"unknown field":   "version: 1\nthings:\n  - name: a\n    colour: red\n",
		"unknown section": "version: 1\nother: []\n",
		"twice":           "version: 1\nthings: []\nthings: []\n",
		"alias":           "version: 1\nthings:\n  - &a {name: a}\n  - *a\n",
		"list root":       "- 1\n",
		"broken":          "version: [1\n",
	} {
		if _, _, err := (YAMLCodec{}).Decode([]byte(file), noEnv, target); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// TestSchemaMatchesSections keeps the published JSON Schema in step with
// the sections the code reads and writes.
func TestSchemaMatchesSections(t *testing.T) {
	data, err := os.ReadFile("../../../../../docs/schema/trisse-home.v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		ID         string                     `json:"$id"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.ID != SchemaURL {
		t.Fatalf("$id %q does not match SchemaURL", schema.ID)
	}
	var keys []string
	for k := range schema.Properties {
		keys = append(keys, k)
	}
	want := append([]string{"version"}, domain.SectionOrder...)
	slices.Sort(keys)
	slices.Sort(want)
	if !slices.Equal(keys, want) {
		t.Fatalf("schema sections %v, code sections %v", keys, want)
	}
}

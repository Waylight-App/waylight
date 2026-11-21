package configfile

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"strings"

	"go.yaml.in/yaml/v3"

	"waylight/internal/modules/dashboard/usecase"
)

// The built-in tab templates (#589) are dashboards in the config-as-code
// format, so a new one is a YAML file, not code.
//
//go:embed templates/*.yaml
var builtinFiles embed.FS

// builtinFile is one template file: a tab of the dashboard section plus
// what the onboarding needs to know about it.
type builtinFile struct {
	ID    string   `yaml:"id"`
	Title string   `yaml:"title"`
	Needs []string `yaml:"needs,omitempty,flow"`
	Tab   Tab      `yaml:"tab"`
}

// Builtin reads the embedded templates in file name order. Unknown fields
// are errors, like in waylight.yaml, so a typo fails the tests.
func Builtin() ([]usecase.BuiltinTemplate, error) {
	names, err := fs.Glob(builtinFiles, "templates/*.yaml")
	if err != nil {
		return nil, err
	}
	out := make([]usecase.BuiltinTemplate, 0, len(names))
	for _, name := range names {
		raw, err := builtinFiles.ReadFile(name)
		if err != nil {
			return nil, err
		}
		dec := yaml.NewDecoder(strings.NewReader(string(raw)))
		dec.KnownFields(true)
		var f builtinFile
		if err := dec.Decode(&f); err != nil {
			return nil, fmt.Errorf("template %s: %w", path.Base(name), err)
		}
		tabs, err := Section{Tabs: []Tab{f.Tab}}.ToTabs()
		if err != nil {
			return nil, fmt.Errorf("template %s: %w", path.Base(name), err)
		}
		out = append(out, usecase.BuiltinTemplate{ID: f.ID, Title: f.Title, Needs: f.Needs, Tab: tabs[0]})
	}
	return out, nil
}

// MergeConfig implements usecase.ConfigMerger: it sets values in a widget's
// JSON settings and keeps the others.
func MergeConfig(config string, values map[string]any) (string, error) {
	cfg := map[string]any{}
	if config != "" {
		if err := json.Unmarshal([]byte(config), &cfg); err != nil {
			return "", errConfig
		}
	}
	maps.Copy(cfg, values)
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", errConfig
	}
	return string(b), nil
}

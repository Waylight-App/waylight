package domain

import (
	"regexp"
	"strings"

	"waylight/internal/shared/kernel"
)

// ConfigVersion is the version of Waylight.yaml (#88). Like exports,
// files of a newer version are rejected.
const ConfigVersion = 1

// MaxConfigBytes limits an uploaded config file.
const MaxConfigBytes = 2 << 20

// SectionOrder is the order of the sections in an exported file and the
// order in which an imported file is applied.
var SectionOrder = []string{"appearance", "search", "services", "connections", "bookmarks", "dashboard"}

// CheckConfigVersion validates the version of a config file.
func CheckConfigVersion(v int) error {
	if v < 1 || v > ConfigVersion {
		return kernel.NewValidationErrorf("version", "Config version {version} is not supported by this version of Waylight.", "version", v)
	}
	return nil
}

// envRef matches ${NAME}; $${NAME} is the escape for a literal ${NAME}.
var envRef = regexp.MustCompile(`\$?\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// ExpandEnv replaces ${NAME} with the value of the environment variable
// NAME, so tokens and private host names can stay in .env instead of the
// file. An unset variable is an error: an empty URL or password would fail
// later and far less clearly.
func ExpandEnv(s string, lookup func(string) (string, bool)) (string, error) {
	if !strings.Contains(s, "${") {
		return s, nil
	}
	var missing string
	out := envRef.ReplaceAllStringFunc(s, func(m string) string {
		if strings.HasPrefix(m, "$$") {
			return m[1:]
		}
		name := envRef.FindStringSubmatch(m)[1]
		v, ok := lookup(name)
		if !ok && missing == "" {
			missing = name
		}
		return v
	})
	if missing != "" {
		return "", kernel.NewValidationErrorf("env", "The environment variable {name} is not set.", "name", missing)
	}
	return out, nil
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"waylight/internal/modules"
	"waylight/internal/platform/config"
)

// postYAML sends a config file and returns the status and body.
func postYAML(t *testing.T, c *client, body string) (int, []byte) {
	t.Helper()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/transfer/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/yaml")
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	return res.StatusCode, out
}

// TestConfigAsCode covers waylight.yaml (#88) end to end: the sample
// data survives export → import → export byte for byte, ${NAME} comes from
// the environment, and mistakes are reported without changing anything.
func TestConfigAsCode(t *testing.T) {
	t.Setenv("WAYLIGHT_TEST_DOMAIN", "example.org")
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.DisabledModules = []string{"weather", "calendar", "waste"} })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	// The first visit creates the Home tab, so the round trip covers tabs too.
	c.do(http.MethodGet, "/api/dashboard", nil, http.StatusOK)
	first := c.do(http.MethodGet, "/api/transfer/config", nil, http.StatusOK)
	if !bytes.Contains(first, []byte("- name: Home")) {
		t.Fatalf("no tab in the export:\n%s", first)
	}
	for _, want := range []string{"$schema=", "version: 1", "appearance:", "search:", "services:", "integrations:", "bookmarks:", "dashboard:"} {
		if !bytes.Contains(first, []byte(want)) {
			t.Fatalf("export lacks %q:\n%s", want, first)
		}
	}
	status, body := postYAML(t, c, string(first))
	var res struct {
		Applied []string          `json:"applied"`
		Failed  map[string]string `json:"failed"`
	}
	_ = json.Unmarshal(body, &res)
	if status != http.StatusOK || len(res.Applied) != 6 || len(res.Failed) != 0 {
		t.Fatalf("re-import: %d %s", status, body)
	}
	if second := c.do(http.MethodGet, "/api/transfer/config", nil, http.StatusOK); !bytes.Equal(first, second) {
		t.Fatalf("round trip changed the file:\n--- first\n%s\n--- second\n%s", first, second)
	}

	own := `version: 1
services:
  - name: Jellyfin
    url: https://jf.${WAYLIGHT_TEST_DOMAIN}
bookmarks:
  - title: Docs
    url: https://docs.example.org
    tags: [work]
  - folder: Media
    bookmarks:
      - { title: Films, url: "https://films.${WAYLIGHT_TEST_DOMAIN}" }
dashboard:
  tabs:
    - name: Home
      rules: [{ days: [1, 2, 3, 4, 5], from: "07:00", to: "09:00" }]
      widgets:
        - kind: clock
          at: { x: 0, y: 0, w: 4, h: 2 }
          phone: { order: 2, rows: 2 }
        - kind: bookmarks
          at: { x: 4, y: 0, w: 4, h: 4 }
          phone: { order: 1, rows: 4, hidden: true }
`
	if status, body := postYAML(t, c, own); status != http.StatusOK || !strings.Contains(string(body), `"failed":{}`) {
		t.Fatalf("own file: %d %s", status, body)
	}
	var services []obj
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, &services)
	if len(services) != 1 || services[0]["url"] != "https://jf.example.org" {
		t.Fatalf("services: %v", services)
	}
	var dash obj
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	tabs := dash["tabs"].([]any)
	if len(tabs) != 1 || len(tabs[0].(obj)["widgets"].([]any)) != 2 {
		t.Fatalf("dashboard: %v", dash)
	}
	again := c.do(http.MethodGet, "/api/transfer/config", nil, http.StatusOK)
	if !bytes.Contains(again, []byte("phone: {order: 1, rows: 4, hidden: true}")) || !bytes.Contains(again, []byte("url: https://jf.example.org")) {
		t.Fatalf("export after own file:\n%s", again)
	}

	// Rejected before anything changes.
	for name, file := range map[string]string{
		"unknown field":    "version: 1\nservices:\n  - name: A\n    url: https://a.example\n    colour: red\n",
		"missing variable": "version: 1\nservices:\n  - name: A\n    url: https://${WAYLIGHT_NOT_SET}\n",
		"unknown section":  "version: 1\nrecipes: []\n",
		"newer version":    "version: 2\n",
		"not yaml":         "version: [1\n",
	} {
		if status, body := postYAML(t, c, file); status != http.StatusUnprocessableEntity {
			t.Errorf("%s: %d %s", name, status, body)
		}
	}
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, &services)
	if len(services) != 1 {
		t.Fatalf("a rejected file changed services: %v", services)
	}
	// A section that fails validation is reported; the others are applied.
	status, body = postYAML(t, c, "version: 1\nservices: []\ndashboard:\n  tabs:\n    - name: Home\n      widgets:\n        - { kind: no-such-widget, at: { x: 0, y: 0, w: 1, h: 1 } }\n")
	_ = json.Unmarshal(body, &res)
	if status != http.StatusOK || res.Failed["dashboard"] == "" || len(res.Applied) != 1 || res.Applied[0] != "services" {
		t.Fatalf("partial: %d %s", status, body)
	}
}

// recordingRouter remembers every registered pattern.
type recordingRouter struct{ patterns map[string]bool }

func (r recordingRouter) Public(p string, _ http.HandlerFunc)    { r.patterns[p] = true }
func (r recordingRouter) Protected(p string, _ http.HandlerFunc) { r.patterns[p] = true }

// TestConfigRoutesAreRegistered keeps ConfigRoutes in step with the routes
// a module really registers; a renamed route would otherwise stay writable
// in managed mode.
func TestConfigRoutesAreRegistered(t *testing.T) {
	up := fakeUpstream(t)
	cfg := newTestConfig(t, up.URL)
	a, err := build(context.Background(), cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer a.db.Close()
	found := 0
	for _, m := range a.registry.All() {
		c, ok := m.(modules.Configurable)
		if !ok {
			continue
		}
		found++
		rec := recordingRouter{patterns: map[string]bool{}}
		m.RegisterRoutes(rec)
		for _, route := range c.ConfigRoutes() {
			if !rec.patterns[route] {
				t.Errorf("%s: config route %q is not registered", m.Name(), route)
			}
		}
	}
	if found != 6 {
		t.Fatalf("expected six configurable modules, found %d", found)
	}
}

// TestManagedMode applies CONFIG_FILE at start: its sections refuse
// changes through the API and uploads, the others stay editable, and a
// broken file stops the start.
func TestManagedMode(t *testing.T) {
	file := filepath.Join(t.TempDir(), "waylight.yaml")
	_ = os.WriteFile(file, []byte("version: 1\nservices:\n  - name: NAS\n    url: https://nas.example\n"), 0o600)
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) {
		c.ConfigFile = file
		c.SampleData = true
		c.DisabledModules = []string{"weather", "calendar", "waste"}
	})
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var services []obj
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, &services)
	if len(services) != 1 || services[0]["name"] != "NAS" {
		t.Fatalf("sample data must not touch a managed section: %v", services)
	}
	c.do(http.MethodPost, "/api/services", obj{"name": "X", "url": "https://x.example"}, http.StatusConflict)
	c.do(http.MethodDelete, "/api/services/"+services[0]["id"].(string), nil, http.StatusConflict)
	c.do(http.MethodPost, "/api/bookmarks", obj{"url": "https://free.example"}, http.StatusCreated)

	var status struct {
		Managed []string `json:"managed"`
	}
	c.json(http.MethodGet, "/api/transfer/config/status", nil, http.StatusOK, &status)
	if len(status.Managed) != 1 || status.Managed[0] != "services" {
		t.Fatalf("status: %+v", status)
	}
	_, body := postYAML(t, c, "version: 1\nservices: []\n")
	if !strings.Contains(string(body), "managed by CONFIG_FILE") {
		t.Fatalf("upload overrode the file: %s", body)
	}

	// A file that fails at start stops the server from starting.
	_ = os.WriteFile(file, []byte("version: 1\nservices:\n  - name: NAS\n    url: ftp://nas\n"), 0o600)
	cfg := newTestConfig(t, up.URL, func(c *config.Config) { c.ConfigFile = file })
	if _, err := build(context.Background(), cfg, slog.New(slog.NewTextHandler(io.Discard, nil))); err == nil || !strings.Contains(err.Error(), "services") {
		t.Fatalf("broken CONFIG_FILE: %v", err)
	}
}

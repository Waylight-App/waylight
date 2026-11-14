package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"trisse-home/internal/platform/config"
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

// TestConfigAsCode covers trisse-home.yaml (#88) end to end: the sample
// data survives export → import → export byte for byte, ${NAME} comes from
// the environment, and mistakes are reported without changing anything.
func TestConfigAsCode(t *testing.T) {
	t.Setenv("TRISSE_TEST_DOMAIN", "example.org")
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
	for _, want := range []string{"$schema=", "version: 1", "appearance:", "search:", "services:", "bookmarks:", "dashboard:"} {
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
	if status != http.StatusOK || len(res.Applied) != 5 || len(res.Failed) != 0 {
		t.Fatalf("re-import: %d %s", status, body)
	}
	if second := c.do(http.MethodGet, "/api/transfer/config", nil, http.StatusOK); !bytes.Equal(first, second) {
		t.Fatalf("round trip changed the file:\n--- first\n%s\n--- second\n%s", first, second)
	}

	own := `version: 1
services:
  - name: Jellyfin
    url: https://jf.${TRISSE_TEST_DOMAIN}
bookmarks:
  - title: Docs
    url: https://docs.example.org
    tags: [work]
  - folder: Media
    bookmarks:
      - { title: Films, url: "https://films.${TRISSE_TEST_DOMAIN}" }
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
		"missing variable": "version: 1\nservices:\n  - name: A\n    url: https://${TRISSE_NOT_SET}\n",
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

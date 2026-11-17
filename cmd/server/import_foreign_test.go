package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"

	"waylight/internal/platform/config"
)

type foreignResult struct {
	YAML    string `json:"yaml"`
	Summary struct {
		Tiles, Bookmarks, Widgets, Skipped int
		Tab                                string
		Warnings                           []string
	} `json:"summary"`
	Result *struct {
		Applied []string          `json:"applied"`
		Failed  map[string]string `json:"failed"`
	} `json:"result"`
}

func importForeign(t *testing.T, c *client, format string, apply bool, paths ...string) foreignResult {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		fw, _ := mw.CreateFormFile("file", p[len("../../internal/modules/transfer/adapter/gateway/testdata/homepage/"):])
		_, _ = fw.Write(data)
	}
	_ = mw.Close()
	url := c.base + "/api/transfer/import/" + format
	if apply {
		url += "?apply=true"
	}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, url, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("import: %d %s", res.StatusCode, out)
	}
	var r foreignResult
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

// TestImportFromHomepage covers #77 end to end: preview, apply, and a
// second import that adds nothing.
func TestImportFromHomepage(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.DisabledModules = []string{"weather", "calendar", "waste"} })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)
	dir := "../../internal/modules/transfer/adapter/gateway/testdata/homepage/"
	files := []string{dir + "services.yaml", dir + "bookmarks.yaml", dir + "widgets.yaml"}

	preview := importForeign(t, c, "homepage", false, files...)
	if preview.Result != nil || preview.Summary.Tiles != 4 || preview.Summary.Bookmarks != 2 || preview.Summary.Tab != "Homepage" {
		t.Fatalf("preview: %+v", preview.Summary)
	}
	var services []obj
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, &services)
	if len(services) != 0 {
		t.Fatalf("a preview must not change anything: %v", services)
	}

	applied := importForeign(t, c, "homepage", true, files...)
	if applied.Result == nil || len(applied.Result.Failed) != 0 {
		t.Fatalf("apply: %+v", applied.Result)
	}
	c.json(http.MethodGet, "/api/services", nil, http.StatusOK, &services)
	if len(services) != 4 {
		t.Fatalf("services: %v", services)
	}
	var dash obj
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	found := false
	for _, tab := range dash["tabs"].([]any) {
		if tab.(obj)["name"] == "Homepage" {
			found = len(tab.(obj)["widgets"].([]any)) == 4 // services, bookmarks, system, clock; weather is disabled
		}
	}
	if !found {
		t.Fatalf("import tab: %v", dash["tabs"])
	}

	again := importForeign(t, c, "homepage", false, files...)
	if again.Summary.Tiles != 0 || again.Summary.Bookmarks != 0 || again.Summary.Skipped != 6 || again.Summary.Tab != "" {
		t.Fatalf("a second import must add nothing: %+v", again.Summary)
	}
}

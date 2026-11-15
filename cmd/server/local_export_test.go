package main

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

// TestLocalModeExportImports checks that a file exported by the browser
// extension's local mode (ADR 0019) imports into a server unchanged. The
// file is written by web/src/app/local/localExport.test.ts, so a change of
// a table on either side fails one of the two tests.
func TestLocalModeExportImports(t *testing.T) {
	raw, err := os.ReadFile("../../web/src/app/local/testdata/local-export.json")
	if err != nil {
		t.Fatal(err)
	}
	var file obj
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var res struct {
		Imported []string          `json:"imported"`
		Skipped  []string          `json:"skipped"`
		Failed   map[string]string `json:"failed"`
	}
	c.json(http.MethodPost, "/api/transfer/import", file, http.StatusOK, &res)
	if len(res.Failed) > 0 || len(res.Skipped) > 0 || len(res.Imported) != 2 {
		t.Fatalf("import result: %+v", res)
	}

	var dash struct {
		Tabs []struct {
			Name         string `json:"name"`
			Background   string `json:"background"`
			MobileCustom bool   `json:"mobileCustom"`
			Rules        []obj  `json:"rules"`
			Widgets      []struct {
				Kind   string `json:"kind"`
				Config obj    `json:"config"`
			} `json:"widgets"`
		} `json:"tabs"`
	}
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	if len(dash.Tabs) != 2 || dash.Tabs[0].Name != "Home" || dash.Tabs[1].Name != "Evening" {
		t.Fatalf("tabs after import: %+v", dash.Tabs)
	}
	home, evening := dash.Tabs[0], dash.Tabs[1]
	if len(home.Widgets) != 1 || home.Widgets[0].Kind != "clock" {
		t.Fatalf("home widgets: %+v", home.Widgets)
	}
	if len(home.Rules) != 1 || home.Rules[0]["from"] != "07:00" || home.Rules[0]["to"] != "09:30" {
		t.Fatalf("home rules: %+v", home.Rules)
	}
	if !evening.MobileCustom || evening.Background == "" || evening.Widgets[0].Config["scope"] != "pinned" {
		t.Fatalf("evening tab: %+v", evening)
	}

	var templates []obj
	c.json(http.MethodGet, "/api/dashboard/templates", nil, http.StatusOK, &templates)
	if len(templates) != 1 || templates[0]["name"] != "Start" {
		t.Fatalf("templates: %v", templates)
	}

	var look obj
	c.json(http.MethodGet, "/api/appearance", nil, http.StatusOK, &look)
	if look["theme"] != "dark" || look["greetingName"] != "Sam" || look["profile"] != "everyday" {
		t.Fatalf("appearance after import: %v", look)
	}
}

package main

import (
	"encoding/json"
	"fmt"
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
	if len(res.Failed) > 0 || len(res.Skipped) > 0 || len(res.Imported) != 8 {
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

	var notes []obj
	c.json(http.MethodGet, "/api/notes", nil, http.StatusOK, &notes)
	if len(notes) != 1 || notes[0]["title"] != "Shopping" || notes[0]["pinned"] != true {
		t.Fatalf("notes after import: %v", notes)
	}
	var revs []obj
	c.json(http.MethodGet, "/api/notes/"+notes[0]["id"].(string)+"/revisions", nil, http.StatusOK, &revs)
	if len(revs) != 1 {
		t.Fatalf("note revisions after import: %v", revs)
	}

	var todos []obj
	c.json(http.MethodGet, "/api/todos?done=true", nil, http.StatusOK, &todos)
	if len(todos) != 2 {
		t.Fatalf("to-dos after import: %v", todos)
	}
	water := todos[0]
	if water["title"] != "Water plants" || water["due"] != "2026-11-18" {
		t.Fatalf("recurring to-do after import: %v", water)
	}
	if rep, _ := water["repeat"].(map[string]any); rep["kind"] != "weekly" {
		t.Fatalf("recurrence after import: %v", water["repeat"])
	}

	var marks struct {
		Bookmarks []obj `json:"bookmarks"`
		Folders   []obj `json:"folders"`
	}
	c.json(http.MethodGet, "/api/bookmarks", nil, http.StatusOK, &marks)
	if len(marks.Bookmarks) != 2 || len(marks.Folders) != 1 || marks.Folders[0]["count"] != float64(1) {
		t.Fatalf("bookmarks after import: %+v", marks)
	}
	if tags, _ := marks.Bookmarks[0]["tags"].([]any); len(tags) != 2 {
		t.Fatalf("bookmark tags after import: %v", marks.Bookmarks[0])
	}

	var habits []obj
	c.json(http.MethodGet, "/api/habits", nil, http.StatusOK, &habits)
	if len(habits) != 1 || habits[0]["name"] != "Read" || fmt.Sprint(habits[0]["days"]) != "[1 3 5]" {
		t.Fatalf("habits after import: %v", habits)
	}

	// Countdowns, focus sessions and completed days are read back through the
	// server's own export: their lists depend on today's date, the rows do not.
	var back struct {
		Sections map[string]map[string][]obj `json:"sections"`
	}
	c.json(http.MethodGet, "/api/transfer/export", nil, http.StatusOK, &back)
	for table, want := range map[[2]string]int{
		{"countdown", "countdown_items"}: 2,
		{"focus", "focus_sessions"}:      1,
		{"habits", "habit_completions"}:  2,
	} {
		if got := len(back.Sections[table[0]][table[1]]); got != want {
			t.Fatalf("%s rows after import: got %d, want %d", table[1], got, want)
		}
	}
	for _, row := range back.Sections["countdown"]["countdown_items"] {
		if row["title"] == "Dentist" && row["minute"] != float64(870) {
			t.Fatalf("timed countdown after import: %v", row)
		}
	}
}

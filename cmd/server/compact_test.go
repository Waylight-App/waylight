package main

import (
	"net/http"
	"strings"
	"testing"
)

// TestCompactStates: the compact state of to-dos, habits and focus (#597)
// and of services, backups, drives and the calendar (#594).
func TestCompactStates(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var state struct {
		Title struct {
			Key, Text string
		} `json:"title"`
		Level string `json:"level"`
		Count *int   `json:"count"`
	}
	c.do(http.MethodPost, "/api/todos", obj{"title": "Tax return"}, http.StatusCreated)
	c.json(http.MethodGet, "/api/todos/compact", nil, http.StatusOK, &state)
	if state.Count == nil || *state.Count < 1 || state.Title.Key == "" || state.Level == "" {
		t.Fatalf("todos: %+v", state)
	}
	c.json(http.MethodGet, "/api/habits/compact", nil, http.StatusOK, &state)
	if state.Title.Text == "" {
		t.Fatalf("habits: %+v", state)
	}
	c.json(http.MethodGet, "/api/focus/compact", nil, http.StatusOK, &state)
	if state.Title.Text != "0 min focus today" || state.Level != "none" {
		t.Fatalf("focus: %+v", state)
	}
	// The focus timer starts and pauses through its compact action (#631).
	var focus struct {
		Title  struct{ Text string } `json:"title"`
		Action struct {
			ID, Method, Path string
		} `json:"action"`
	}
	for _, want := range []string{"focus.start", "focus.pause", "focus.start"} {
		c.json(http.MethodGet, "/api/focus/compact", nil, http.StatusOK, &focus)
		if focus.Action.ID != want {
			t.Fatalf("focus action %+v, want %s", focus, want)
		}
		c.do(focus.Action.Method, focus.Action.Path, nil, http.StatusOK)
	}
	var timer struct {
		Running bool   `json:"running"`
		Phase   string `json:"phase"`
	}
	c.json(http.MethodGet, "/api/focus/timer", nil, http.StatusOK, &timer)
	if !timer.Running || timer.Phase != "work" {
		t.Fatalf("timer after start, pause, start: %+v", timer)
	}
	c.json(http.MethodGet, "/api/focus/compact", nil, http.StatusOK, &focus)
	if !strings.HasPrefix(focus.Title.Text, "Focus · ") {
		t.Fatalf("running title %q", focus.Title.Text)
	}
	for _, path := range []string{"/api/services/compact", "/api/backupmonitor/compact", "/api/diskhealth/compact", "/api/calendar/compact"} {
		state.Title.Text, state.Level = "", ""
		c.json(http.MethodGet, path, nil, http.StatusOK, &state)
		if state.Title.Key == "" || state.Title.Text == "" || state.Level == "" {
			t.Fatalf("%s: %+v", path, state)
		}
	}
}

package main

import (
	"net/http"
	"testing"
)

// TestCompactStates: the compact state of to-dos, habits and focus (#597).
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
}

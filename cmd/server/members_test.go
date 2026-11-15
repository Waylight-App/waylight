package main

import (
	"net/http"
	"testing"
)

// TestMembersStayOff: until the personal-content modules store owners
// (ADR 0018), the server refuses to create members.
func TestMembersStayOff(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	var list obj
	owner.json(http.MethodGet, "/api/auth/members", nil, http.StatusOK, &list)
	if list["available"] != false {
		t.Fatalf("members must be off: %v", list)
	}
	owner.do(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusConflict)
}

// TestMemberAccess checks the guard for members (#294) with members switched
// on: they change personal content and settings, read instance data, and
// manage nothing of the instance.
func TestMemberAccess(t *testing.T) {
	saved := membersEnabled
	membersEnabled = true
	t.Cleanup(func() { membersEnabled = saved })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var member obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusCreated, &member)
	var group obj
	owner.json(http.MethodPost, "/api/auth/groups", obj{"name": "Office", "accounts": []string{member["id"].(string)}}, http.StatusCreated, &group)

	m := newClient(t, srv.URL)
	var ms obj
	m.json(http.MethodPost, "/api/auth/login", obj{"username": "ben", "password": "member-password-1"}, http.StatusOK, &ms)
	m.csrf = ms["csrfToken"].(string)
	var status obj
	m.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &status)
	if status["role"] != "member" || status["username"] != "ben" {
		t.Fatalf("status: %v", status)
	}

	// Personal content and settings: read and write.
	m.do(http.MethodGet, "/api/notes", nil, http.StatusOK)
	m.do(http.MethodPost, "/api/notes", obj{"title": "Mine", "body": "x"}, http.StatusCreated)
	m.do(http.MethodGet, "/api/appearance", nil, http.StatusOK)
	m.do(http.MethodGet, "/api/modules", nil, http.StatusOK)

	// Instance data: read only; instance settings and admin areas: closed.
	m.do(http.MethodGet, "/api/services", nil, http.StatusOK)
	m.do(http.MethodPost, "/api/services", obj{"name": "x", "url": "http://x"}, http.StatusForbidden)
	for _, path := range []string{"/api/auth/guests", "/api/auth/members", "/api/auth/groups", "/api/auth/events", "/api/auth/tokens",
		"/api/transfer/export", "/api/backups", "/api/notifications/settings", "/api/adguard/settings", "/api/docker/containers"} {
		m.do(http.MethodGet, path, nil, http.StatusForbidden)
	}
	m.do(http.MethodPost, "/api/auth/members", obj{"name": "eve", "password": "member-password-2"}, http.StatusForbidden)
	m.do(http.MethodPost, "/api/auth/guests", obj{"name": "eve", "password": "guest-password-2"}, http.StatusForbidden)

	// Self-service on the own account works.
	m.do(http.MethodGet, "/api/auth/sessions", nil, http.StatusOK)

	// The security log records the account and group changes.
	var events []obj
	owner.json(http.MethodGet, "/api/auth/events", nil, http.StatusOK, &events)
	seen := map[string]bool{}
	for _, e := range events {
		seen[e["type"].(string)] = true
	}
	if !seen["member_changed"] || !seen["group_changed"] {
		t.Fatalf("events: %v", seen)
	}

	// Disabling signs the member out at once.
	owner.do(http.MethodPut, "/api/auth/members/"+member["id"].(string), obj{"disabled": true}, http.StatusOK)
	m.do(http.MethodGet, "/api/notes", nil, http.StatusUnauthorized)
	owner.do(http.MethodDelete, "/api/auth/members/"+member["id"].(string), nil, http.StatusNoContent)
	owner.do(http.MethodDelete, "/api/auth/groups/"+group["id"].(string), nil, http.StatusNoContent)
}

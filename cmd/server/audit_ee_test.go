//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"waylight/ee/licence/domain"
)

// TestAuditLog checks #303 in the ee build: changes to instance data and
// sharing are recorded without content, a member's private notes are not,
// admins read and export the log, and nobody can change an entry.
func TestAuditLog(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved, savedMembers := licencePublicKey, membersEnabled
	licencePublicKey, membersEnabled = pub, true
	t.Cleanup(func() { licencePublicKey, membersEnabled = saved, savedMembers })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	// Without a licence key the log neither records nor answers.
	owner.do(http.MethodPost, "/api/services", obj{"name": "Before", "url": up.URL + "/health"}, http.StatusCreated)
	owner.do(http.MethodGet, "/api/audit", nil, http.StatusForbidden)

	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"audit-log"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	owner.do(http.MethodPost, "/api/services", obj{"name": "NAS", "url": up.URL + "/health"}, http.StatusCreated)

	var m obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusCreated, &m)
	ben := newClient(t, srv.URL)
	var bs obj
	ben.json(http.MethodPost, "/api/auth/login", obj{"username": "ben", "password": "member-password-1"}, http.StatusOK, &bs)
	ben.csrf = bs["csrfToken"].(string)
	var note obj
	ben.json(http.MethodPost, "/api/notes", obj{"title": "Private plan", "body": "secret"}, http.StatusCreated, &note)
	ben.do(http.MethodPut, "/api/notes/"+note["id"].(string)+"/visibility", obj{"visibility": "everyone"}, http.StatusOK)
	ben.do(http.MethodGet, "/api/audit", nil, http.StatusForbidden)

	var entries []obj
	owner.json(http.MethodGet, "/api/audit", nil, http.StatusOK, &entries)
	has := func(route, account string) bool {
		for _, e := range entries {
			if e["route"] == route && e["accountId"] == account {
				return true
			}
		}
		return false
	}
	if !has("POST /api/services", "owner") || !has("PUT /api/notes/{id}/visibility", m["id"].(string)) {
		t.Fatalf("expected entries missing: %v", entries)
	}
	if has("POST /api/notes", m["id"].(string)) {
		t.Fatalf("a member's private note was logged: %v", entries)
	}
	// Only one service entry: the one created before the key was not recorded.
	services := 0
	for _, e := range entries {
		if e["route"] == "POST /api/services" {
			services++
		}
	}
	if services != 1 {
		t.Fatalf("service entries: %d", services)
	}
	res, err := owner.http.Get(srv.URL + "/api/audit/export?format=csv")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK || !strings.HasPrefix(string(body), "at,account_id,role,module,route,object_id,id") ||
		!strings.Contains(string(body), "POST /api/services") {
		t.Fatalf("csv export: %d %s", res.StatusCode, body)
	}

	owner.do(http.MethodPut, "/api/audit/settings", obj{"retentionDays": 7}, http.StatusUnprocessableEntity)
	owner.do(http.MethodPut, "/api/audit/settings", obj{"retentionDays": 90}, http.StatusOK)
	// The change of the retention is itself in the log.
	owner.json(http.MethodGet, "/api/audit?module=audit", nil, http.StatusOK, &entries)
	if len(entries) != 1 || entries[0]["route"] != "PUT /api/audit/settings" {
		t.Fatalf("retention change not logged: %v", entries)
	}
	// There is no way to change or delete an entry.
	owner.do(http.MethodDelete, "/api/audit/"+entries[0]["id"].(string), nil, http.StatusNotFound)
}

//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"slices"
	"testing"
	"time"

	"waylight/ee/licence/domain"
)

// TestPresence checks #346 in the ee build: own statuses, planned absences
// winning, the switch and the licence gate.
func TestPresence(t *testing.T) {
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
	owner.do(http.MethodGet, "/api/presence", nil, http.StatusForbidden)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"presence", "absences"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	var modulesList struct {
		WidgetKinds []string `json:"widgetKinds"`
	}
	owner.json(http.MethodGet, "/api/modules", nil, http.StatusOK, &modulesList)
	if !slices.Contains(modulesList.WidgetKinds, "presence") {
		t.Fatalf("widget kinds: %v", modulesList.WidgetKinds)
	}

	var anna, ben obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &anna)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated, &ben)
	a := newClient(t, srv.URL)
	var s obj
	a.json(http.MethodPost, "/api/auth/login", obj{"username": "anna", "password": "member-password-1"}, http.StatusOK, &s)
	a.csrf = s["csrfToken"].(string)

	today := time.Now().Format(time.DateOnly)
	owner.do(http.MethodPost, "/api/absences", obj{"accountId": ben["id"], "from": today, "to": today, "kind": "remote"}, http.StatusCreated)
	a.do(http.MethodPut, "/api/presence/me", obj{"status": "office"}, http.StatusNoContent)
	a.do(http.MethodPut, "/api/presence/me", obj{"status": "lunch"}, http.StatusUnprocessableEntity)
	a.do(http.MethodPut, "/api/presence/settings", obj{"enabled": false}, http.StatusForbidden)

	var board struct {
		Day    string
		People []obj
		Counts map[string]int
		Mine   obj
	}
	a.json(http.MethodGet, "/api/presence", nil, http.StatusOK, &board)
	if board.Day != today || len(board.People) != 3 || board.Counts["office"] != 1 || board.Counts["home"] != 1 ||
		board.Counts["none"] != 1 || board.Mine["status"] != "office" {
		t.Fatalf("board: %+v", board)
	}
	if p := board.People[1]; p["name"] != "ben" || p["status"] != "home" || p["source"] != "absence" {
		t.Fatalf("ben: %v", p)
	}
	a.do(http.MethodGet, "/api/presence?group=nope", nil, http.StatusUnprocessableEntity)

	// Switched off, presence answers 404 and forgets today's statuses.
	owner.do(http.MethodPut, "/api/presence/settings", obj{"enabled": false}, http.StatusOK)
	a.do(http.MethodGet, "/api/presence", nil, http.StatusNotFound)
	owner.do(http.MethodPut, "/api/presence/settings", obj{"enabled": true}, http.StatusOK)
	a.json(http.MethodGet, "/api/presence", nil, http.StatusOK, &board)
	if board.Mine["status"] != "" {
		t.Fatalf("status survived switching off: %v", board.Mine)
	}
}

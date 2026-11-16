//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"trisse-home/ee/licence/domain"
)

// TestAdminPolicies checks #304 in the ee build.
func TestAdminPolicies(t *testing.T) {
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

	owner.do(http.MethodPut, "/api/policies", obj{"minPasswordLength": 16}, http.StatusForbidden)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"policies"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	owner.do(http.MethodPut, "/api/policies", obj{"minPasswordLength": 8}, http.StatusUnprocessableEntity)
	owner.do(http.MethodPut, "/api/policies", obj{"minPasswordLength": 16, "memberModules": []string{"notes"}}, http.StatusOK)

	owner.do(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "only-12-char"}, http.StatusUnprocessableEntity)
	owner.do(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "sixteen-chars-ok"}, http.StatusCreated)
	ben := newClient(t, srv.URL)
	var bs obj
	ben.json(http.MethodPost, "/api/auth/login", obj{"username": "ben", "password": "sixteen-chars-ok"}, http.StatusOK, &bs)
	ben.csrf = bs["csrfToken"].(string)

	var mods struct {
		Modules []string `json:"modules"`
	}
	ben.json(http.MethodGet, "/api/modules", nil, http.StatusOK, &mods)
	if slices.Contains(mods.Modules, "todos") || !slices.Contains(mods.Modules, "notes") || !slices.Contains(mods.Modules, "dashboard") {
		t.Fatalf("member modules: %v", mods.Modules)
	}
	ben.do(http.MethodGet, "/api/todos", nil, http.StatusForbidden)
	ben.do(http.MethodGet, "/api/notes", nil, http.StatusOK)
	owner.json(http.MethodGet, "/api/modules", nil, http.StatusOK, &mods)
	if !slices.Contains(mods.Modules, "todos") {
		t.Fatal("the owner lost a module")
	}

	// A required second factor: ben has none and may only set one up.
	owner.do(http.MethodPut, "/api/policies", obj{"requireSecondFactor": true}, http.StatusOK)
	var st obj
	ben.json(http.MethodGet, "/api/auth/status", nil, http.StatusOK, &st)
	if st["secondFactorRequired"] != true {
		t.Fatalf("status: %v", st)
	}
	ben.do(http.MethodGet, "/api/notes", nil, http.StatusForbidden)

	// Policy changes are in the security log.
	body := owner.do(http.MethodGet, "/api/auth/events", nil, http.StatusOK)
	if !strings.Contains(string(body), "policies_changed") {
		t.Fatalf("security log: %s", body)
	}
}

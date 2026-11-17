//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"testing"
	"time"

	"waylight/ee/licence/domain"
)

// TestLicenceKeys runs the licence endpoints of the ee build (#307) against a
// real server, with a test key pair instead of the compiled-in public key.
func TestLicenceKeys(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	saved, savedMembers := licencePublicKey, membersEnabled
	licencePublicKey, membersEnabled = pub, true
	t.Cleanup(func() { licencePublicKey, membersEnabled = saved, savedMembers })
	key := func(seats int, features ...string) string {
		now := time.Now().UTC()
		return domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0),
			Seats: seats, Features: features}, priv)
	}

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var lic obj
	owner.json(http.MethodGet, "/api/licence", nil, http.StatusOK, &lic)
	if lic["status"] != "missing" || lic["accounts"] != float64(1) {
		t.Fatalf("fresh instance: %v", lic)
	}
	var features struct {
		Features map[string]string `json:"features"`
	}
	owner.json(http.MethodGet, "/api/licence/features", nil, http.StatusOK, &features)
	if features.Features["branding"] != "none" {
		t.Fatalf("features without key: %v", features)
	}

	owner.do(http.MethodPut, "/api/licence", obj{"key": "THL1.not-a-key"}, http.StatusUnprocessableEntity)
	owner.json(http.MethodPut, "/api/licence", obj{"key": key(2, "branding")}, http.StatusOK, &lic)
	if lic["status"] != "valid" || lic["licensee"] != "Example GmbH" || lic["source"] != "settings" {
		t.Fatalf("installed: %v", lic)
	}
	owner.json(http.MethodGet, "/api/licence/features", nil, http.StatusOK, &features)
	if features.Features["branding"] != "full" || features.Features["sso"] != "none" {
		t.Fatalf("features with key: %v", features)
	}

	// Two seats: owner plus one member fit, a guest is one account too many.
	var m obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusCreated, &m)
	member := newClient(t, srv.URL)
	var ms obj
	member.json(http.MethodPost, "/api/auth/login", obj{"username": "ben", "password": "member-password-1"}, http.StatusOK, &ms)
	member.csrf = ms["csrfToken"].(string)
	member.do(http.MethodGet, "/api/licence", nil, http.StatusForbidden)
	member.json(http.MethodGet, "/api/licence/features", nil, http.StatusOK, &features)
	if features.Features["branding"] != "full" {
		t.Fatalf("member within seats: %v", features)
	}
	owner.do(http.MethodPost, "/api/auth/guests", obj{"name": "anna", "password": "guest-password-1", "tabs": []string{}}, http.StatusCreated)
	owner.json(http.MethodGet, "/api/licence", nil, http.StatusOK, &lic)
	if lic["status"] != "over-seats" || lic["accounts"] != float64(3) {
		t.Fatalf("over seats: %v", lic)
	}
	member.json(http.MethodGet, "/api/licence/features", nil, http.StatusOK, &features)
	if features.Features["branding"] != "read" {
		t.Fatalf("member over seats: %v", features)
	}

	owner.do(http.MethodDelete, "/api/licence", nil, http.StatusNoContent)
	owner.json(http.MethodGet, "/api/licence", nil, http.StatusOK, &lic)
	if lic["status"] != "missing" {
		t.Fatalf("after removal: %v", lic)
	}
}

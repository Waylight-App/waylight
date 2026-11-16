//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"slices"
	"testing"
	"time"

	"trisse-home/ee/licence/domain"
)

// TestAbsences checks #299 in the ee build: own and admin-entered absences,
// private reasons, the range and the licence gate.
func TestAbsences(t *testing.T) {
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
	owner.do(http.MethodGet, "/api/absences?from=2026-12-01&to=2026-12-31", nil, http.StatusForbidden)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"absences"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)
	// The widget of the ee/ plugin (#299) can be placed once the feature is licensed.
	var info struct {
		WidgetKinds []string `json:"widgetKinds"`
	}
	owner.json(http.MethodGet, "/api/modules", nil, http.StatusOK, &info)
	if !slices.Contains(info.WidgetKinds, "absences") {
		t.Fatalf("widget kinds: %v", info.WidgetKinds)
	}

	var anna, ben obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &anna)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated, &ben)
	login := func(name, pw string) *client {
		c := newClient(t, srv.URL)
		var s obj
		c.json(http.MethodPost, "/api/auth/login", obj{"username": name, "password": pw}, http.StatusOK, &s)
		c.csrf = s["csrfToken"].(string)
		return c
	}
	a, b := login("anna", "member-password-1"), login("ben", "member-password-2")

	var sick, holiday obj
	a.json(http.MethodPost, "/api/absences", obj{"from": "2026-12-01", "to": "2026-12-02", "kind": "sick", "note": "doctor", "private": true},
		http.StatusCreated, &sick)
	a.do(http.MethodPost, "/api/absences", obj{"accountId": ben["id"], "from": "2026-12-03", "to": "2026-12-03", "kind": "holiday"},
		http.StatusForbidden)
	owner.json(http.MethodPost, "/api/absences", obj{"accountId": ben["id"], "from": "2026-12-20", "to": "2027-01-03", "kind": "holiday"},
		http.StatusCreated, &holiday)
	owner.do(http.MethodPost, "/api/absences", obj{"accountId": "nobody", "from": "2026-12-03", "to": "2026-12-03", "kind": "holiday"},
		http.StatusUnprocessableEntity)

	var list []obj
	b.json(http.MethodGet, "/api/absences?from=2026-12-01&to=2026-12-31", nil, http.StatusOK, &list)
	if len(list) != 2 || list[0]["name"] != "anna" || list[0]["kind"] != "away" || list[0]["note"] != "" || list[0]["editable"] != false ||
		list[1]["kind"] != "holiday" || list[1]["editable"] != true {
		t.Fatalf("ben's view: %v", list)
	}
	owner.json(http.MethodGet, "/api/absences?from=2026-12-01&to=2026-12-31", nil, http.StatusOK, &list)
	if list[0]["kind"] != "sick" || list[0]["note"] != "doctor" {
		t.Fatalf("admin's view: %v", list)
	}
	b.do(http.MethodPut, "/api/absences/"+sick["id"].(string), obj{"from": "2026-12-01", "to": "2026-12-05", "kind": "holiday"}, http.StatusForbidden)
	b.do(http.MethodDelete, "/api/absences/"+holiday["id"].(string), nil, http.StatusNoContent)
	b.do(http.MethodGet, "/api/absences?from=2026-01-01&to=2026-12-31", nil, http.StatusUnprocessableEntity)

	// A deleted member's absences go with the account.
	owner.do(http.MethodDelete, "/api/auth/members/"+anna["id"].(string), nil, http.StatusNoContent)
	owner.json(http.MethodGet, "/api/absences?from=2026-12-01&to=2026-12-31", nil, http.StatusOK, &list)
	if len(list) != 0 {
		t.Fatalf("after deleting anna: %v", list)
	}
}

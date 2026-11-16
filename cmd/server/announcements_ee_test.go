//go:build ee

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"testing"
	"time"

	"trisse-home/ee/licence/domain"
)

// TestAnnouncements checks #297 in the ee build: who writes, who reads,
// read marks, expiry into the archive, and the licence gate.
func TestAnnouncements(t *testing.T) {
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

	// Without a key the module is closed.
	owner.do(http.MethodGet, "/api/announcements", nil, http.StatusForbidden)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"announcements"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	var anna, ben obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &anna)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated, &ben)
	var office obj
	owner.json(http.MethodPost, "/api/auth/groups", obj{"name": "Office", "accounts": []string{anna["id"].(string)}}, http.StatusCreated, &office)
	login := func(name, pw string) *client {
		c := newClient(t, srv.URL)
		var s obj
		c.json(http.MethodPost, "/api/auth/login", obj{"username": name, "password": pw}, http.StatusOK, &s)
		c.csrf = s["csrfToken"].(string)
		return c
	}
	a, b := login("anna", "member-password-1"), login("ben", "member-password-2")

	// Admins write; members only once they are in the writer group.
	var closed obj
	owner.json(http.MethodPost, "/api/announcements", obj{"title": "Office closed on Friday", "body": "**All day.**", "pinned": true},
		http.StatusCreated, &closed)
	var me obj
	a.json(http.MethodGet, "/api/announcements/me", nil, http.StatusOK, &me)
	if me["mayWrite"] != false {
		t.Fatalf("anna before the writer group: %v", me)
	}
	owner.json(http.MethodGet, "/api/announcements/me", nil, http.StatusOK, &me)
	if me["mayWrite"] != true {
		t.Fatalf("owner: %v", me)
	}
	a.do(http.MethodPost, "/api/announcements", obj{"title": "Cake"}, http.StatusForbidden)
	a.do(http.MethodGet, "/api/announcements/settings", nil, http.StatusForbidden)
	owner.do(http.MethodPut, "/api/announcements/settings", obj{"writerGroup": office["id"]}, http.StatusOK)
	a.json(http.MethodGet, "/api/announcements/me", nil, http.StatusOK, &me)
	if me["mayWrite"] != true {
		t.Fatalf("anna in the writer group: %v", me)
	}
	var cake obj
	a.json(http.MethodPost, "/api/announcements", obj{"title": "Cake in the kitchen"}, http.StatusCreated, &cake)
	b.do(http.MethodPost, "/api/announcements", obj{"title": "Not me"}, http.StatusForbidden)
	// Writers change their own notices only; admins change all.
	a.do(http.MethodPut, "/api/announcements/"+closed["id"].(string), obj{"title": "Hijacked"}, http.StatusForbidden)
	owner.do(http.MethodPut, "/api/announcements/"+cake["id"].(string), obj{"title": "Cake in the kitchen (vegan)"}, http.StatusOK)
	a.do(http.MethodPost, "/api/announcements", obj{"title": ""}, http.StatusUnprocessableEntity)

	// Everyone reads; pinned first; read marks are per person.
	var list []obj
	b.json(http.MethodGet, "/api/announcements", nil, http.StatusOK, &list)
	if len(list) != 2 || list[0]["id"] != closed["id"] || list[0]["read"] != false || list[0]["editable"] != false ||
		list[0]["authorName"] != "owner" || list[1]["authorName"] != "anna" {
		t.Fatalf("ben's list: %v", list)
	}
	b.do(http.MethodPost, "/api/announcements/"+closed["id"].(string)+"/read", nil, http.StatusNoContent)
	b.json(http.MethodGet, "/api/announcements", nil, http.StatusOK, &list)
	if list[0]["read"] != true {
		t.Fatalf("not marked read: %v", list[0])
	}
	a.json(http.MethodGet, "/api/announcements", nil, http.StatusOK, &list)
	if list[0]["read"] != false || list[1]["editable"] != true {
		t.Fatalf("anna's list: %v", list)
	}

	// An expired notice leaves the widget for the archive.
	past := now.Add(-time.Hour).Format(time.RFC3339)
	owner.do(http.MethodPut, "/api/announcements/"+cake["id"].(string), obj{"title": "Cake was yesterday", "expiresAt": past}, http.StatusOK)
	b.json(http.MethodGet, "/api/announcements", nil, http.StatusOK, &list)
	if len(list) != 1 {
		t.Fatalf("current after expiry: %v", list)
	}
	b.json(http.MethodGet, "/api/announcements?archive=1", nil, http.StatusOK, &list)
	if len(list) != 1 || list[0]["id"] != cake["id"] {
		t.Fatalf("archive: %v", list)
	}

	// Posting to Slack or Teams needs chat apps in the licence (#332).
	var channels []obj
	owner.json(http.MethodGet, "/api/announcements/channels", nil, http.StatusOK, &channels)
	if len(channels) != 0 {
		t.Fatalf("channels without chat apps: %v", channels)
	}
	b.do(http.MethodGet, "/api/announcements/channels", nil, http.StatusForbidden)
	owner.do(http.MethodPost, "/api/announcements", obj{"title": "To Slack", "postTo": "c1"}, http.StatusForbidden)

	// A deleted notice is gone, with its read marks.
	owner.do(http.MethodDelete, "/api/announcements/"+closed["id"].(string), nil, http.StatusNoContent)
	b.do(http.MethodPost, "/api/announcements/"+closed["id"].(string)+"/read", nil, http.StatusNotFound)
}

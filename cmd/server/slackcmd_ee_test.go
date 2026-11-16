//go:build ee

package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSlackFinderSearchesWithTheMembersRights checks the start page search
// behind the Slack command (#333): a member finds services, their own
// bookmarks and shared ones, never another member's private bookmark.
func TestSlackFinderSearchesWithTheMembersRights(t *testing.T) {
	savedMembers := membersEnabled
	membersEnabled = true
	t.Cleanup(func() { membersEnabled = savedMembers })

	up := fakeUpstream(t)
	a, err := build(context.Background(), newTestConfig(t, up.URL), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.db.Close() })
	srv := httptest.NewServer(newHandler(a))
	t.Cleanup(srv.Close)

	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	var anna, ben obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &anna)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated, &ben)
	owner.do(http.MethodPost, "/api/services", obj{"name": "Office VPN", "url": "https://vpn.example"}, http.StatusCreated)
	login := func(name, pw string) *client {
		c := newClient(t, srv.URL)
		var s obj
		c.json(http.MethodPost, "/api/auth/login", obj{"username": name, "password": pw}, http.StatusOK, &s)
		c.csrf = s["csrfToken"].(string)
		return c
	}
	login("ben", "member-password-2").do(http.MethodPost, "/api/bookmarks", obj{"title": "Ben's VPN notes", "url": "https://ben.example/vpn"},
		http.StatusCreated)
	login("anna", "member-password-1").do(http.MethodPost, "/api/bookmarks", obj{"title": "VPN client download", "url": "https://dl.example/vpn"},
		http.StatusCreated)

	hits, err := startPageFinder{a.registry}.Find(context.Background(), anna["id"].(string), "vpn")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, h := range hits {
		got[h.Title] = h.Source
	}
	if got["Office VPN"] != "Service" || got["VPN client download"] != "Bookmark" {
		t.Fatalf("anna's hits: %v", hits)
	}
	if _, leaked := got["Ben's VPN notes"]; leaked {
		t.Fatalf("anna finds ben's private bookmark: %v", hits)
	}
}

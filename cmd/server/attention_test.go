package main

import (
	"net/http"
	"testing"
)

// TestAttentionList: a departures widget without a stop and an unconfigured
// Immich widget (and the other integrations that report a status) show up in GET /api/attention; choosing a stop removes the
// first entry (#599).
func TestAttentionList(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	var tab, transit obj
	owner.json(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Check"}, http.StatusCreated, &tab)
	tabID := tab["id"].(string)
	owner.json(http.MethodPost, "/api/dashboard/tabs/"+tabID+"/widgets", obj{"kind": "transit", "x": 0, "y": 0, "w": 4, "h": 2}, http.StatusCreated, &transit)
	// Integrations that report their connection (#599, #592), none set up.
	for i, kind := range []string{"immich", "media", "arr", "proxmox", "speedtest"} {
		owner.do(http.MethodPost, "/api/dashboard/tabs/"+tabID+"/widgets", obj{"kind": kind, "x": 0, "y": 2 + 2*i, "w": 4, "h": 2}, http.StatusCreated)
	}

	var list struct {
		Items []obj `json:"items"`
		Count int   `json:"count"`
	}
	owner.json(http.MethodGet, "/api/attention", nil, http.StatusOK, &list)
	kinds := map[string]string{}
	for _, it := range list.Items {
		kinds[it["integration"].(string)] = it["action"].(string)
	}
	if kinds["transit"] != "choose" || list.Count != len(list.Items) {
		t.Fatalf("attention: %+v", list)
	}
	for _, kind := range []string{"immich", "media", "arr", "proxmox", "speedtest"} {
		if kinds[kind] != "connect" {
			t.Fatalf("%s is not set up: %+v", kind, list)
		}
	}

	// Connections & services lists the same integrations with their state (#591).
	var conns []obj
	owner.json(http.MethodGet, "/api/attention/connections", nil, http.StatusOK, &conns)
	states := map[string]string{}
	for _, c := range conns {
		states[c["integration"].(string)] = c["state"].(string)
	}
	if states["immich"] != "setup" || states["proxmox"] != "setup" {
		t.Fatalf("connections: %+v", conns)
	}

	owner.json(http.MethodPatch, "/api/dashboard/widgets/"+transit["id"].(string), obj{"config": obj{"stopId": "de-delfi:1"}}, http.StatusOK, nil)
	owner.json(http.MethodGet, "/api/attention", nil, http.StatusOK, &list)
	for _, it := range list.Items {
		if it["integration"] == "transit" {
			t.Fatalf("the stop was chosen: %+v", list.Items)
		}
	}
}

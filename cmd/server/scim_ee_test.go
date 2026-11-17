//go:build ee

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"waylight/ee/licence/domain"
)

// TestSCIMProvisioning runs a provider's typical SCIM sequence against the ee
// build (#396): look up, create, group, deactivate with Entra's string
// "False", delete.
func TestSCIMProvisioning(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved, savedMembers := licencePublicKey, membersEnabled
	licencePublicKey, membersEnabled = pub, true
	t.Cleanup(func() { licencePublicKey, membersEnabled = saved, savedMembers })
	t.Setenv("SCIM_TOKEN", "scim-secret-token")

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	scim := func(method, path, token string, body any, want int) obj {
		t.Helper()
		var buf bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&buf).Encode(body)
		}
		req, _ := http.NewRequest(method, srv.URL+path, &buf)
		req.Header.Set("Content-Type", "application/scim+json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var out obj
		_ = json.NewDecoder(res.Body).Decode(&out)
		if res.StatusCode != want {
			t.Fatalf("%s %s: status %d, want %d: %v", method, path, res.StatusCode, want, out)
		}
		return out
	}
	const tok = "scim-secret-token"

	// Without a licence key that includes directory-sync the routes are closed.
	scim(http.MethodGet, "/scim/v2/Users", tok, nil, http.StatusForbidden)
	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 10,
		Features: []string{"directory-sync"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	scim(http.MethodGet, "/scim/v2/Users", "", nil, http.StatusUnauthorized)
	scim(http.MethodGet, "/scim/v2/Users", "wrong", nil, http.StatusUnauthorized)

	list := scim(http.MethodGet, `/scim/v2/Users?filter=userName+eq+%22anna@example.org%22`, tok, nil, http.StatusOK)
	if list["totalResults"] != float64(0) {
		t.Fatalf("lookup before create: %v", list)
	}
	anna := scim(http.MethodPost, "/scim/v2/Users", tok, obj{"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"userName": "anna@example.org", "externalId": "entra-1", "active": true, "displayName": "Anna Weber"}, http.StatusCreated)
	annaID := anna["id"].(string)
	if anna["userName"] != "anna@example.org" || anna["active"] != true {
		t.Fatalf("created: %v", anna)
	}
	scim(http.MethodPost, "/scim/v2/Users", tok, obj{"userName": "ANNA@example.org"}, http.StatusConflict)
	ben := scim(http.MethodPost, "/scim/v2/Users", tok, obj{"userName": "ben@example.org"}, http.StatusCreated)
	benID := ben["id"].(string)

	office := scim(http.MethodPost, "/scim/v2/Groups", tok, obj{"displayName": "Office", "members": []obj{{"value": annaID}}}, http.StatusCreated)
	officeID := office["id"].(string)
	scim(http.MethodPatch, "/scim/v2/Groups/"+officeID, tok, obj{"Operations": []obj{
		{"op": "Add", "path": "members", "value": []obj{{"value": benID}}},
		{"op": "Remove", "path": `members[value eq "` + annaID + `"]`},
	}}, http.StatusOK)
	got := scim(http.MethodGet, "/scim/v2/Groups/"+officeID, tok, nil, http.StatusOK)
	if m := got["members"].([]any); len(m) != 1 || m[0].(map[string]any)["value"] != benID {
		t.Fatalf("group members: %v", got)
	}

	// Entra ID deactivates with the string "False".
	scim(http.MethodPatch, "/scim/v2/Users/"+annaID, tok, obj{"Operations": []obj{{"op": "Replace", "path": "active", "value": "False"}}}, http.StatusOK)
	var members struct {
		Members []obj `json:"members"`
	}
	owner.json(http.MethodGet, "/api/auth/members", nil, http.StatusOK, &members)
	for _, m := range members.Members {
		if m["id"] == annaID && m["disabled"] != true {
			t.Fatalf("anna not disabled: %v", m)
		}
	}

	// Okta sends the values without a path.
	benOut := scim(http.MethodPatch, "/scim/v2/Users/"+benID, tok, obj{"Operations": []obj{{"op": "replace", "value": obj{"active": false, "displayName": "Ben"}}}}, http.StatusOK)
	if benOut["active"] != false || benOut["displayName"] != "Ben" {
		t.Fatalf("path-less patch: %v", benOut)
	}

	scim(http.MethodDelete, "/scim/v2/Users/"+annaID, tok, nil, http.StatusNoContent)
	scim(http.MethodGet, "/scim/v2/Users/"+annaID, tok, nil, http.StatusNotFound)
	scim(http.MethodDelete, "/scim/v2/Groups/"+officeID, tok, nil, http.StatusNoContent)
	scim(http.MethodGet, "/scim/v2/ServiceProviderConfig", tok, nil, http.StatusOK)
}

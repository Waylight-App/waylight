package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestShoppingListsOnTrisseHome checks #471 through the API: lists without
// CalDAV, sharing with a group, rights of owner and others, the widget's
// current list, import from CalDAV and export.
func TestShoppingListsOnTrisseHome(t *testing.T) {
	saved := membersEnabled
	membersEnabled = true
	t.Cleanup(func() { membersEnabled = saved })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	var anna, ben, carl obj
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "anna", "password": "member-password-1"}, http.StatusCreated, &anna)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-2"}, http.StatusCreated, &ben)
	owner.json(http.MethodPost, "/api/auth/members", obj{"name": "carl", "password": "member-password-3"}, http.StatusCreated, &carl)
	var family obj
	owner.json(http.MethodPost, "/api/auth/groups", obj{"name": "Family", "accounts": []string{anna["id"].(string), ben["id"].(string)}},
		http.StatusCreated, &family)
	login := func(name, pw string) *client {
		c := newClient(t, srv.URL)
		var s obj
		c.json(http.MethodPost, "/api/auth/login", obj{"username": name, "password": pw}, http.StatusOK, &s)
		c.csrf = s["csrfToken"].(string)
		return c
	}
	a, b, c := login("anna", "member-password-1"), login("ben", "member-password-2"), login("carl", "member-password-3")

	// Without any CalDAV settings: create a list and use it.
	var lists obj
	a.json(http.MethodGet, "/api/shopping/lists", nil, http.StatusOK, &lists)
	if len(lists["lists"].([]any)) != 0 || lists["caldav"] != false || lists["current"] != "" {
		t.Fatalf("empty start: %v", lists)
	}
	var groceries obj
	a.json(http.MethodPost, "/api/shopping/lists", obj{"name": "  Groceries  "}, http.StatusCreated, &groceries)
	if groceries["name"] != "Groceries" || groceries["visibility"] != "private" || groceries["mine"] != true {
		t.Fatalf("new list: %v", groceries)
	}
	id := groceries["id"].(string)
	var milk obj
	a.json(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Milk, 1 l; organic"}, http.StatusCreated, &milk)
	a.do(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Bread"}, http.StatusCreated)
	a.do(http.MethodPut, "/api/shopping/current", obj{"list": id}, http.StatusNoContent)

	// Private: nobody else reaches it.
	b.do(http.MethodGet, "/api/shopping/lists/"+id+"/items", nil, http.StatusNotFound)
	b.do(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Chocolate"}, http.StatusNotFound)

	// Shared with Family: ben adds and ticks off, carl still sees nothing.
	a.do(http.MethodPut, "/api/shopping/lists/"+id, obj{"name": "Groceries", "visibility": "group:" + family["id"].(string)}, http.StatusOK)
	var choc obj
	b.json(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Chocolate"}, http.StatusCreated, &choc)
	b.do(http.MethodPut, "/api/shopping/lists/"+id+"/items/"+milk["id"].(string), obj{"done": true}, http.StatusNoContent)
	c.do(http.MethodGet, "/api/shopping/lists/"+id+"/items", nil, http.StatusNotFound)
	c.do(http.MethodDelete, "/api/shopping/lists/"+id+"/items/"+choc["id"].(string), nil, http.StatusNotFound)
	var bensView obj
	b.json(http.MethodGet, "/api/shopping/lists", nil, http.StatusOK, &bensView)
	shared := bensView["lists"].([]any)
	if len(shared) != 1 || shared[0].(map[string]any)["mine"] != false || shared[0].(map[string]any)["open"] != float64(2) {
		t.Fatalf("ben's lists: %v", bensView)
	}
	// Only the owner renames, shares or deletes.
	b.do(http.MethodPut, "/api/shopping/lists/"+id, obj{"name": "Mine now", "visibility": "everyone"}, http.StatusForbidden)
	b.do(http.MethodDelete, "/api/shopping/lists/"+id, nil, http.StatusForbidden)
	a.do(http.MethodPut, "/api/shopping/lists/"+id, obj{"name": "Groceries", "visibility": "group:not-mine"}, http.StatusUnprocessableEntity)

	var items []obj
	a.json(http.MethodGet, "/api/shopping/lists/"+id+"/items", nil, http.StatusOK, &items)
	if len(items) != 3 || items[len(items)-1]["done"] != true {
		t.Fatalf("items, open first: %v", items)
	}

	// Export as text and as iCalendar.
	get := func(path string) (string, http.Header) {
		res, err := a.http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)
		if res.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d %s", path, res.StatusCode, body)
		}
		return string(body), res.Header
	}
	text, hdr := get("/api/shopping/lists/" + id + "/export?format=text")
	if !strings.Contains(text, "[ ] Bread") || !strings.Contains(text, "[x] Milk, 1 l; organic") ||
		!strings.Contains(hdr.Get("Content-Disposition"), `filename=Groceries.txt`) {
		t.Fatalf("text export: %q %v", text, hdr)
	}
	ics, _ := get("/api/shopping/lists/" + id + "/export?format=ics")
	if !strings.Contains(ics, "BEGIN:VTODO") || !strings.Contains(ics, `SUMMARY:Milk\, 1 l\; organic`) ||
		!strings.Contains(ics, "STATUS:COMPLETED") || !strings.Contains(ics, "\r\n") {
		t.Fatalf("ics export: %q", ics)
	}

	// Import the open items of a CalDAV list once.
	a.do(http.MethodPost, "/api/shopping/lists/"+id+"/import", nil, http.StatusNotFound) // no CalDAV yet
	a.json(http.MethodPut, "/api/shopping/settings", obj{"listUrl": up.URL + "/dav/tasks/", "username": "me", "password": "pw"}, http.StatusOK, nil)
	a.do(http.MethodPost, "/api/shopping/items", obj{"summary": "Coffee"}, http.StatusCreated)
	var imported obj
	a.json(http.MethodPost, "/api/shopping/lists/"+id+"/import", nil, http.StatusOK, &imported)
	if imported["imported"] == float64(0) {
		t.Fatalf("import: %v", imported)
	}
	a.json(http.MethodGet, "/api/shopping/lists", nil, http.StatusOK, &lists)
	if lists["caldav"] != true || lists["current"] != id {
		t.Fatalf("after connecting CalDAV: %v", lists)
	}

	// Clear ticked items, then delete the list.
	var cleared obj
	b.json(http.MethodPost, "/api/shopping/lists/"+id+"/clear-completed", nil, http.StatusOK, &cleared)
	if cleared["deleted"] != float64(1) {
		t.Fatalf("cleared: %v", cleared)
	}
	a.do(http.MethodDelete, "/api/shopping/lists/"+id, nil, http.StatusNoContent)
	b.do(http.MethodGet, "/api/shopping/lists/"+id+"/items", nil, http.StatusNotFound)
}

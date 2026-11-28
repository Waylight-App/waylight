package main

import (
	"net/http"
	"testing"
)

// TestFamilyTemplateSharing checks #871 through the API: applying the Family
// template with a group sets up one shopping list for the whole group and
// the board's default visibility; group members add and tick off items,
// a guest in the group only reads, and an account outside the group sees
// nothing, neither list nor notes.
func TestFamilyTemplateSharing(t *testing.T) {
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
	login := func(name, pw string) *client {
		c := newClient(t, srv.URL)
		var s obj
		c.json(http.MethodPost, "/api/auth/login", obj{"username": name, "password": pw}, http.StatusOK, &s)
		c.csrf = s["csrfToken"].(string)
		return c
	}
	a, b, c := login("anna", "member-password-1"), login("ben", "member-password-2"), login("carl", "member-password-3")

	// The guest is in the group and sees the owner's Family tab.
	var ownerTab obj
	owner.json(http.MethodPost, "/api/dashboard/templates/builtin/family/tabs", obj{"name": "Family"}, http.StatusCreated, &ownerTab)
	var gina obj
	owner.json(http.MethodPost, "/api/auth/guests", obj{"name": "gina", "password": "guest-password-1", "tabs": []string{ownerTab["id"].(string)}},
		http.StatusCreated, &gina)
	g := login("gina", "guest-password-1")
	var family obj
	owner.json(http.MethodPost, "/api/auth/groups",
		obj{"name": "Family", "accounts": []string{anna["id"].(string), ben["id"].(string), gina["id"].(string)}}, http.StatusCreated, &family)
	share := "group:" + family["id"].(string)

	// Carl may not share with a group he is not in; nothing is created.
	c.do(http.MethodPost, "/api/dashboard/templates/builtin/family/tabs", obj{"name": "Family", "visibility": share}, http.StatusUnprocessableEntity)
	var carlLists obj
	c.json(http.MethodGet, "/api/shopping/lists", nil, http.StatusOK, &carlLists)
	if len(carlLists["lists"].([]any)) != 0 {
		t.Fatalf("carl's lists after a refused share: %v", carlLists)
	}

	// Anna applies the template: one list for the group, current for her,
	// and the board widget defaults to the group.
	var annaTab obj
	a.json(http.MethodPost, "/api/dashboard/templates/builtin/family/tabs", obj{"name": "Family", "visibility": share}, http.StatusCreated, &annaTab)
	boardShared := false
	for _, w := range annaTab["widgets"].([]any) {
		wm := w.(map[string]any)
		if cfg, _ := wm["config"].(map[string]any); wm["kind"] == "board" && cfg["visibility"] == share {
			boardShared = true
		}
	}
	if !boardShared {
		t.Fatalf("board widget without the group: %v", annaTab["widgets"])
	}
	var lists obj
	a.json(http.MethodGet, "/api/shopping/lists", nil, http.StatusOK, &lists)
	if len(lists["lists"].([]any)) != 1 {
		t.Fatalf("anna's lists: %v", lists)
	}
	list := lists["lists"].([]any)[0].(map[string]any)
	id := list["id"].(string)
	if list["name"] != "Family" || list["visibility"] != share || list["mine"] != true || lists["current"] != id {
		t.Fatalf("shared list: %v", lists)
	}

	// Ben applies it too and lands on the same list, not a second one.
	b.do(http.MethodPost, "/api/dashboard/templates/builtin/family/tabs", obj{"name": "Family", "visibility": share}, http.StatusCreated)
	var bens obj
	b.json(http.MethodGet, "/api/shopping/lists", nil, http.StatusOK, &bens)
	if len(bens["lists"].([]any)) != 1 || bens["current"] != id || bens["lists"].([]any)[0].(map[string]any)["mine"] != false {
		t.Fatalf("ben's lists: %v", bens)
	}

	// Members of the group add, tick off and remove items.
	var milk, bread obj
	b.json(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Milk"}, http.StatusCreated, &milk)
	a.json(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Bread"}, http.StatusCreated, &bread)
	a.do(http.MethodPut, "/api/shopping/lists/"+id+"/items/"+milk["id"].(string), obj{"done": true}, http.StatusNoContent)
	b.do(http.MethodDelete, "/api/shopping/lists/"+id+"/items/"+bread["id"].(string), nil, http.StatusNoContent)
	// Only the owner renames, shares or deletes the list.
	b.do(http.MethodPut, "/api/shopping/lists/"+id, obj{"name": "Ben's", "visibility": share}, http.StatusForbidden)
	b.do(http.MethodDelete, "/api/shopping/lists/"+id, nil, http.StatusForbidden)

	// The guest in the group reads the list but changes nothing.
	var items []obj
	g.json(http.MethodGet, "/api/shopping/lists/"+id+"/items", nil, http.StatusOK, &items)
	if len(items) != 1 || items[0]["summary"] != "Milk" {
		t.Fatalf("guest's items: %v", items)
	}
	g.do(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Sweets"}, http.StatusForbidden)
	g.do(http.MethodPut, "/api/shopping/lists/"+id+"/items/"+milk["id"].(string), obj{"done": false}, http.StatusForbidden)
	g.do(http.MethodDelete, "/api/shopping/lists/"+id+"/items/"+milk["id"].(string), nil, http.StatusForbidden)

	// Carl, outside the group, reaches nothing.
	c.do(http.MethodGet, "/api/shopping/lists/"+id+"/items", nil, http.StatusNotFound)
	c.do(http.MethodPost, "/api/shopping/lists/"+id+"/items", obj{"summary": "Beer"}, http.StatusNotFound)
	c.do(http.MethodPut, "/api/shopping/lists/"+id+"/items/"+milk["id"].(string), obj{"done": false}, http.StatusNotFound)
	c.do(http.MethodDelete, "/api/shopping/lists/"+id+"/items/"+milk["id"].(string), nil, http.StatusNotFound)
	c.do(http.MethodPut, "/api/shopping/current", obj{"list": id}, http.StatusNotFound)

	// Board: a note for the group is read by members and the guest, changed
	// only by its writer, and hidden from carl.
	var note obj
	a.json(http.MethodPost, "/api/board", obj{"text": "Parents' evening on Tuesday", "visibility": share}, http.StatusCreated, &note)
	nid := note["id"].(string)
	seen := func(cl *client) bool {
		var notes []obj
		cl.json(http.MethodGet, "/api/board", nil, http.StatusOK, &notes)
		for _, n := range notes {
			if n["id"] == nid {
				return true
			}
		}
		return false
	}
	if !seen(b) || !seen(g) || seen(c) {
		t.Fatal("board note: ben and gina see it, carl does not")
	}
	b.do(http.MethodPut, "/api/board/"+nid, obj{"text": "Changed"}, http.StatusForbidden)
	g.do(http.MethodPut, "/api/board/"+nid, obj{"text": "Changed"}, http.StatusForbidden)
	c.do(http.MethodPut, "/api/board/"+nid, obj{"text": "Changed"}, http.StatusNotFound)
	b.do(http.MethodDelete, "/api/board/"+nid, nil, http.StatusForbidden)
	// Deleting is idempotent: a note carl cannot see answers like a missing
	// one, and it stays.
	c.do(http.MethodDelete, "/api/board/"+nid, nil, http.StatusNoContent)
	if !seen(a) {
		t.Fatal("anna's note was removed by someone else")
	}
}

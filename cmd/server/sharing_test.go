package main

import (
	"net/http"
	"testing"
)

// TestSharing checks #301 within ADR 0018: an owner shares a note or a
// to-do with a group; group members read it, others do not, and nobody but
// the owner changes it or its visibility.
func TestSharing(t *testing.T) {
	saved := membersEnabled
	membersEnabled = true
	t.Cleanup(func() { membersEnabled = saved })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)

	signIn := func(name string) (*client, string) {
		var m obj
		owner.json(http.MethodPost, "/api/auth/members", obj{"name": name, "password": "member-password-1"}, http.StatusCreated, &m)
		c := newClient(t, srv.URL)
		var s obj
		c.json(http.MethodPost, "/api/auth/login", obj{"username": name, "password": "member-password-1"}, http.StatusOK, &s)
		c.csrf = s["csrfToken"].(string)
		return c, m["id"].(string)
	}
	anna, annaID := signIn("anna")
	ben, benID := signIn("ben")
	carla, _ := signIn("carla")
	var office obj
	owner.json(http.MethodPost, "/api/auth/groups", obj{"name": "Office", "accounts": []string{annaID, benID}}, http.StatusCreated, &office)
	officeVis := "group:" + office["id"].(string)

	var mine []obj
	anna.json(http.MethodGet, "/api/auth/my-groups", nil, http.StatusOK, &mine)
	if len(mine) != 1 || mine[0]["name"] != "Office" {
		t.Fatalf("anna's groups: %v", mine)
	}
	carla.json(http.MethodGet, "/api/auth/my-groups", nil, http.StatusOK, &mine)
	if len(mine) != 0 {
		t.Fatalf("carla is in no group: %v", mine)
	}

	var note obj
	anna.json(http.MethodPost, "/api/notes", obj{"title": "Plan", "body": "for the office"}, http.StatusCreated, &note)
	id := note["id"].(string)
	if note["visibility"] != "private" || note["mine"] != true {
		t.Fatalf("a member's new note: %v", note)
	}
	ben.do(http.MethodGet, "/api/notes/"+id, nil, http.StatusNotFound)
	anna.do(http.MethodPut, "/api/notes/"+id+"/visibility", obj{"visibility": "group:nonexistent"}, http.StatusUnprocessableEntity)
	var shared obj
	anna.json(http.MethodPut, "/api/notes/"+id+"/visibility", obj{"visibility": officeVis}, http.StatusOK, &shared)
	if shared["visibility"] != officeVis {
		t.Fatalf("shared note: %v", shared)
	}
	var seen obj
	ben.json(http.MethodGet, "/api/notes/"+id, nil, http.StatusOK, &seen)
	if seen["mine"] != false {
		t.Fatalf("ben must not own anna's note: %v", seen)
	}
	carla.do(http.MethodGet, "/api/notes/"+id, nil, http.StatusNotFound)
	owner.do(http.MethodGet, "/api/notes/"+id, nil, http.StatusNotFound) // admins see only what is shared with them
	ben.do(http.MethodPut, "/api/notes/"+id, obj{"title": "Mine", "body": "x"}, http.StatusForbidden)
	ben.do(http.MethodPut, "/api/notes/"+id+"/visibility", obj{"visibility": "everyone"}, http.StatusForbidden)

	var todo obj
	anna.json(http.MethodPost, "/api/todos", obj{"title": "Order paper"}, http.StatusCreated, &todo)
	tid := todo["id"].(string)
	anna.do(http.MethodPut, "/api/todos/"+tid+"/visibility", obj{"visibility": officeVis}, http.StatusOK)
	var todos []obj
	ben.json(http.MethodGet, "/api/todos", nil, http.StatusOK, &todos)
	found := false
	for _, td := range todos {
		found = found || td["id"] == tid
	}
	if !found {
		t.Fatal("ben does not see the shared to-do")
	}
	carla.json(http.MethodGet, "/api/todos", nil, http.StatusOK, &todos)
	for _, td := range todos {
		if td["id"] == tid {
			t.Fatal("carla sees a to-do shared with a group she is not in")
		}
	}

	// Back to private: ben loses it again.
	anna.do(http.MethodPut, "/api/notes/"+id+"/visibility", obj{"visibility": "private"}, http.StatusOK)
	ben.do(http.MethodGet, "/api/notes/"+id, nil, http.StatusNotFound)
}

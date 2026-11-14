package main

import (
	"net/http"
	"strings"

	"trisse-home/internal/shared/caldav/caldavtest"
)

// External to-do lists (#169): connect a CalDAV task list, use it from a
// widget's point of view and see changes made in another app.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "todo-sources",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			dav := caldavtest.New(t)
			base := "/api/todos/settings/sources"

			// Discovery from the server's DAV root; a wrong password is refused.
			var found []obj
			c.json(http.MethodPost, base+"/discover", obj{"kind": "caldav", "url": dav.URL + caldavtest.Root, "username": dav.User, "password": dav.Password}, http.StatusOK, &found)
			if len(found) != 1 || found[0]["url"] != dav.ListURL() || found[0]["name"] != "Aufgaben" {
				t.Fatalf("discover: %v", found)
			}
			c.do(http.MethodPost, base, obj{"name": "NC", "kind": "caldav", "url": dav.ListURL(), "username": dav.User, "password": "wrong"}, http.StatusBadGateway)

			var src obj
			c.json(http.MethodPost, base, obj{"name": "Nextcloud", "kind": "caldav", "url": dav.ListURL(), "username": dav.User, "password": dav.Password}, http.StatusCreated, &src)
			id := src["id"].(string)
			if _, leaked := src["password"]; leaked || src["capabilities"].(map[string]any)["repeat"] != false {
				t.Fatalf("source: %v", src)
			}
			q := "?source=" + id

			var todo obj
			c.json(http.MethodPost, "/api/todos"+q, obj{"title": "Steuer", "due": "2026-11-30", "id": "0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70"}, http.StatusCreated, &todo)
			if todo["id"] != "0b7e9d2c-4f1a-4c8e-9a51-2d3c4b5a6f70.ics" || len(dav.Names()) != 1 {
				t.Fatalf("created: %v %v", todo, dav.Names())
			}
			var local []obj
			c.json(http.MethodGet, "/api/todos?done=true", nil, http.StatusOK, &local)
			for _, l := range local {
				if l["title"] == "Steuer" {
					t.Fatal("an external to-do landed in the local list")
				}
			}
			c.do(http.MethodPost, "/api/todos"+q, obj{"title": "x", "due": "2026-11-30", "repeat": obj{"kind": "daily", "every": 1}}, http.StatusUnprocessableEntity)

			// Another app renames it: the next list shows it, an edit of the old version is refused.
			name := dav.Names()[0]
			dav.Put(name, strings.Replace(dav.Data(name), "SUMMARY:Steuer", "SUMMARY:Steuer (Partner)", 1))
			var list []obj
			c.json(http.MethodGet, "/api/todos"+q, nil, http.StatusOK, &list)
			if len(list) != 1 || list[0]["title"] != "Steuer (Partner)" || list[0]["due"] != "2026-11-30" {
				t.Fatalf("list: %v", list)
			}
			c.do(http.MethodPut, "/api/todos/"+name+q, obj{"title": "Mine", "due": "2026-11-30", "done": false, "version": todo["version"]}, http.StatusConflict)
			var done obj
			c.json(http.MethodPut, "/api/todos/"+name+q, obj{"title": "Steuer (Partner)", "due": "2026-11-30", "done": true, "version": list[0]["version"]}, http.StatusOK, &done)
			if done["done"] != true || !strings.Contains(dav.Data(name), "STATUS:COMPLETED") {
				t.Fatalf("done: %v", done)
			}
			c.do(http.MethodPut, "/api/todos/..%2Fsecret.ics"+q, obj{"title": "x", "done": true}, http.StatusNotFound)
			c.json(http.MethodPost, "/api/todos/clear-completed"+q, nil, http.StatusOK, nil)
			if len(dav.Names()) != 0 {
				t.Fatalf("clear completed: %v", dav.Names())
			}

			// Editing without a password keeps it; deleting the list leaves the server alone.
			c.json(http.MethodPut, base+"/"+id, obj{"name": "Cloud", "kind": "caldav", "url": dav.ListURL(), "username": dav.User}, http.StatusOK, &src)
			var all []obj
			c.json(http.MethodGet, base, nil, http.StatusOK, &all)
			if len(all) != 1 || all[0]["name"] != "Cloud" {
				t.Fatalf("sources: %v", all)
			}
			c.do(http.MethodDelete, base+"/"+id, nil, http.StatusNoContent)
			c.do(http.MethodGet, "/api/todos"+q, nil, http.StatusNotFound)
		},
	})
}

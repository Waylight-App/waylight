package main

import "net/http"

// End-to-end check of the places module against the logged-in test server
// (#194, #591): a saved place names the widget that uses it.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "places",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var created obj
			c.json(http.MethodPost, "/api/places", obj{"name": "Home", "label": "Oldenburg", "lat": 53.14, "lon": 8.21}, http.StatusCreated, &created)
			id := created["id"].(string)
			var tab obj
			c.json(http.MethodPost, "/api/dashboard/tabs", obj{"name": "Places"}, http.StatusCreated, &tab)
			c.json(http.MethodPost, "/api/dashboard/tabs/"+tab["id"].(string)+"/widgets",
				obj{"kind": "weather", "x": 0, "y": 0, "w": 4, "h": 2, "config": obj{"placeId": id, "lat": 53.14, "lon": 8.21, "name": "Home"}}, http.StatusCreated, nil)
			var list []obj
			c.json(http.MethodGet, "/api/places", nil, http.StatusOK, &list)
			if len(list) != 1 || list[0]["name"] != "Home" || len(list[0]["usedBy"].([]any)) != 1 {
				t.Fatalf("places: %v", list)
			}
			c.json(http.MethodPut, "/api/places/"+id, obj{"name": "Home sweet home", "lat": 53.14, "lon": 8.21}, http.StatusOK, nil)
			c.do(http.MethodDelete, "/api/places/"+id, nil, http.StatusNoContent)
			c.do(http.MethodDelete, "/api/dashboard/tabs/"+tab["id"].(string), nil, http.StatusNoContent)
		},
	})
}

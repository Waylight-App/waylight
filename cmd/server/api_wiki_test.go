package main

import (
	"fmt"
	"net/http"
	"strings"
)

// End-to-end check of the wiki search (#321): connect a fake BookStack,
// search through it, keep the token secret, and survive a broken wiki.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "wiki",
		Upstream: func(mux *http.ServeMux) {
			mux.HandleFunc("GET /bookstack/api/search", func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Token id:secret" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				base := "http://" + r.Host + "/bookstack"
				fmt.Fprintf(w, `{"data":[{"name":"VPN setup","type":"page","url":%q,"preview_html":{"content":"office <strong>VPN</strong>"},"book":{"name":"IT"}}]}`,
					base+"/books/it/page/vpn")
			})
		},
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var w obj
			c.json(http.MethodPost, "/api/wiki/wikis", obj{"name": "Handbook", "kind": "bookstack", "url": e.up.URL + "/bookstack", "token": "id:secret"}, http.StatusCreated, &w)
			if w["hasToken"] != true || w["token"] != nil {
				t.Fatalf("wiki: %v", w)
			}
			c.do(http.MethodPost, "/api/wiki/wikis/"+w["id"].(string)+"/test", nil, http.StatusNoContent)
			var res obj
			c.json(http.MethodGet, "/api/wiki/search?q=vpn", nil, http.StatusOK, &res)
			pages := res["pages"].([]any)
			if len(pages) != 1 || pages[0].(obj)["space"] != "IT" || pages[0].(obj)["excerpt"] != "office VPN" || !strings.HasSuffix(pages[0].(obj)["url"].(string), "/books/it/page/vpn") {
				t.Fatalf("search: %v", res)
			}
			// A second wiki with a wrong token is named, the first still answers.
			c.do(http.MethodPost, "/api/wiki/wikis", obj{"name": "Broken", "kind": "bookstack", "url": e.up.URL + "/bookstack", "token": "id:wrong"}, http.StatusCreated)
			c.json(http.MethodGet, "/api/wiki/search?q=vpn+again", nil, http.StatusOK, &res)
			if len(res["pages"].([]any)) != 1 || len(res["failed"].([]any)) != 1 {
				t.Fatalf("partial: %v", res)
			}
			c.do(http.MethodPost, "/api/wiki/wikis", obj{"name": "x", "kind": "confluence", "url": "https://x.example", "token": "t"}, http.StatusUnprocessableEntity)
		},
	})
}

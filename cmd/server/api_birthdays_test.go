package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

func init() {
	registerModuleE2E(moduleE2E{
		Name: "birthdays",
		Upstream: func(mux *http.ServeMux) {
			// A CardDAV address book with one contact whose birthday is today.
			mux.HandleFunc("/carddav/contacts/", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "REPORT" {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				bday := time.Now().Format("2006") + time.Now().Format("-01-02")
				w.WriteHeader(http.StatusMultiStatus)
				_, _ = io.WriteString(w, `<d:multistatus xmlns:d="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav"><d:response><d:href>/carddav/contacts/c1.vcf</d:href>`+
					`<d:propstat><d:prop><card:address-data>BEGIN:VCARD&#13;&#10;UID:c1&#13;&#10;FN:Carla&#13;&#10;BDAY:--`+bday[5:]+`&#13;&#10;END:VCARD&#13;&#10;</card:address-data></d:prop>`+
					`<d:status>HTTP/1.1 200 OK</d:status></d:propstat></d:response></d:multistatus>`)
			})
		},
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			c.do(http.MethodPost, "/api/birthdays", obj{"name": "Anna", "month": 2, "day": 30}, http.StatusUnprocessableEntity)
			c.do(http.MethodPost, "/api/birthdays", obj{"name": "Anna", "month": 2, "day": 29, "year": 1992}, http.StatusCreated)

			// vCard import: Anna is already there, Bernd is new.
			vcf := "BEGIN:VCARD\r\nFN:Anna\r\nBDAY:1992-02-29\r\nEND:VCARD\r\nBEGIN:VCARD\r\nFN:Bernd\r\nBDAY:--12-01\r\nEND:VCARD\r\n"
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/birthdays/import", bytes.NewBufferString(vcf))
			req.Header.Set("Content-Type", "text/vcard")
			req.Header.Set("X-CSRF-Token", c.csrf)
			res, err := c.http.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode != http.StatusOK || !strings.Contains(string(body), `"added":1`) || !strings.Contains(string(body), `"duplicates":1`) {
				t.Fatalf("import: %d %s", res.StatusCode, body)
			}

			var s obj
			c.json(http.MethodPut, "/api/birthdays/settings", obj{"remindEnabled": true, "remindAt": "08:00",
				"cardDavUrl": e.up.URL + "/carddav/contacts", "username": "u", "password": "secret"}, http.StatusOK, &s)
			if s["hasPassword"] != true || s["password"] != nil {
				t.Fatalf("settings: %v", s)
			}
			var list []obj
			c.json(http.MethodGet, "/api/birthdays", nil, http.StatusOK, &list)
			if len(list) != 3 || list[0]["name"] != "Carla" || list[0]["inDays"] != float64(0) || list[0]["source"] != "carddav" {
				t.Fatalf("list: %v", list)
			}
			c.do(http.MethodDelete, "/api/birthdays/"+list[0]["id"].(string), nil, http.StatusUnprocessableEntity)
			c.json(http.MethodGet, "/api/birthdays?limit=1", nil, http.StatusOK, &list)
			if len(list) != 1 {
				t.Fatalf("limit: %v", list)
			}
		},
	})
}

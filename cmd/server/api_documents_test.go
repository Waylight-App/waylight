package main

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// putFile sends a file as multipart form data, like the browser does.
func putFile(c *client, id string, data []byte, want int) {
	c.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "receipt.pdf")
	_, _ = fw.Write(data)
	_ = mw.Close()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/documents/"+id+"/file", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	if out, _ := io.ReadAll(res.Body); res.StatusCode != want {
		c.t.Fatalf("upload: %d %s", res.StatusCode, out)
	}
}

// End-to-end check of the documents module against the logged-in test
// server (#688): enter, attach, download, compact view, remove.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "documents",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			end := time.Now().AddDate(0, 0, 10).Format(time.DateOnly)
			var created obj
			c.json(http.MethodPost, "/api/documents", obj{"title": "Kettle", "warrantyEnd": end}, http.StatusCreated, &created)
			id, _ := created["id"].(string)
			if created["status"] != "ending" || created["fileName"] != "" {
				t.Fatalf("documents: %v", created)
			}
			putFile(c, id, []byte("<html>"), http.StatusUnprocessableEntity)
			putFile(c, id, []byte("%PDF-1.4\n"), http.StatusOK)
			if got := c.do(http.MethodGet, "/api/documents/"+id+"/file", nil, http.StatusOK); string(got) != "%PDF-1.4\n" {
				t.Fatalf("download: %q", got)
			}
			var compact obj
			c.json(http.MethodGet, "/api/documents/compact", nil, http.StatusOK, &compact)
			if compact["level"] != "attention" {
				t.Fatalf("documents compact: %v", compact)
			}
			c.do(http.MethodDelete, "/api/documents/"+id, nil, http.StatusNoContent)
			c.do(http.MethodGet, "/api/documents/"+id+"/file", nil, http.StatusNotFound)
		},
	})
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
)

// uploadPhoto posts a picture as multipart form data, like the browser does.
func uploadPhoto(t *testing.T, c *client, album string, data []byte, want int) obj {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "picture.png")
	_, _ = fw.Write(data)
	_ = mw.Close()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/photos?album="+album, &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != want {
		t.Fatalf("upload: %d %s", res.StatusCode, out)
	}
	var o obj
	_ = json.Unmarshal(out, &o)
	return o
}

// End-to-end check of the photo frame (#286): upload, scale, order, serve, delete.
func init() {
	registerModuleE2E(moduleE2E{
		Name: "photos",
		Check: func(e e2eEnv) {
			t, c := e.t, e.c
			var big bytes.Buffer
			_ = png.Encode(&big, image.NewRGBA(image.Rect(0, 0, 3000, 1500)))
			a := uploadPhoto(t, c, "Summer", big.Bytes(), http.StatusCreated)
			if a["width"] != float64(1920) || a["height"] != float64(960) {
				t.Fatalf("not scaled down: %v", a)
			}
			b := uploadPhoto(t, c, "Summer", big.Bytes(), http.StatusCreated)
			uploadPhoto(t, c, "Summer", []byte("<svg/>"), http.StatusUnprocessableEntity)

			c.do(http.MethodPut, "/api/photos/order", obj{"album": "Summer", "ids": []string{b["id"].(string), a["id"].(string)}}, http.StatusNoContent)
			var list []obj
			c.json(http.MethodGet, "/api/photos?album=Summer", nil, http.StatusOK, &list)
			if len(list) != 2 || list[0]["id"] != b["id"] {
				t.Fatalf("order: %v", list)
			}
			var o obj
			c.json(http.MethodGet, "/api/photos/albums", nil, http.StatusOK, &o)
			if o["count"] != float64(2) || o["usedBytes"].(float64) <= 0 || len(o["albums"].([]any)) != 1 {
				t.Fatalf("albums: %v", o)
			}
			if img := c.do(http.MethodGet, "/api/photos/"+a["id"].(string)+"/image?thumb=1", nil, http.StatusOK); len(img) < 3 || img[0] != 0xFF || img[1] != 0xD8 {
				t.Fatal("thumbnail is not a JPEG")
			}
			c.do(http.MethodDelete, "/api/photos/"+a["id"].(string), nil, http.StatusNoContent)
			c.do(http.MethodGet, "/api/photos/"+a["id"].(string)+"/image", nil, http.StatusNotFound)
		},
	})
}

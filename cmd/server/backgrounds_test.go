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

	"trisse-home/internal/platform/config"
)

// TestBackgroundPicker covers #277 end to end: upload, global background,
// a built-in background for one tab, and deleting the picture in use.
func TestBackgroundPicker(t *testing.T) {
	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL, func(c *config.Config) { c.DisabledModules = []string{"weather", "calendar", "waste"} })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf = session["csrfToken"].(string)

	var img bytes.Buffer
	_ = png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 3200, 1800)))
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "beach.png")
	_, _ = fw.Write(img.Bytes())
	_ = mw.Close()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/appearance/backgrounds", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(res.Body)
	res.Body.Close()
	var bg obj
	_ = json.Unmarshal(raw, &bg)
	if res.StatusCode != http.StatusCreated || bg["width"] != float64(2560) {
		t.Fatalf("upload: %d %s", res.StatusCode, raw)
	}
	if jpg := c.do(http.MethodGet, bg["thumb"].(string), nil, http.StatusOK); len(jpg) < 2 || jpg[0] != 0xFF {
		t.Fatal("thumbnail is not a JPEG")
	}

	var a obj
	c.json(http.MethodPut, "/api/appearance/background", obj{"backgroundImage": bg["url"]}, http.StatusOK, &a)
	if a["backgroundImage"] != bg["url"] {
		t.Fatalf("global: %v", a)
	}
	c.do(http.MethodPut, "/api/appearance/background", obj{"backgroundImage": "url(x)"}, http.StatusUnprocessableEntity)

	var dash obj
	c.json(http.MethodGet, "/api/dashboard", nil, http.StatusOK, &dash)
	tab := dash["tabs"].([]any)[0].(obj)["id"].(string)
	var updated obj
	c.json(http.MethodPut, "/api/dashboard/tabs/"+tab+"/background", obj{"background": "builtin:dusk"}, http.StatusOK, &updated)
	if updated["background"] != "builtin:dusk" {
		t.Fatalf("tab: %v", updated)
	}

	c.do(http.MethodDelete, "/api/appearance/backgrounds/"+bg["id"].(string), nil, http.StatusNoContent)
	c.json(http.MethodGet, "/api/appearance", nil, http.StatusOK, &a)
	if a["backgroundImage"] != "" {
		t.Fatalf("a deleted picture must not stay the background: %v", a)
	}
}

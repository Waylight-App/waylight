//go:build ee

package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"trisse-home/ee/licence/domain"
)

// TestBranding checks #305 in the ee build: unlicensed, licensed, and the
// stored branding surviving a removed key.
func TestBranding(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	saved, savedMembers := licencePublicKey, membersEnabled
	licencePublicKey, membersEnabled = pub, true
	t.Cleanup(func() { licencePublicKey, membersEnabled = saved, savedMembers })

	up := fakeUpstream(t)
	srv := newTestServer(t, up.URL)
	owner := newClient(t, srv.URL)
	var session obj
	owner.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	owner.csrf = session["csrfToken"].(string)
	public := newClient(t, srv.URL)
	info := func() obj {
		var o obj
		public.json(http.MethodGet, "/api/branding", nil, http.StatusOK, &o)
		return o
	}
	if b := info(); b["name"] != "trisse-home" || b["custom"] != false {
		t.Fatalf("product branding: %v", b)
	}
	owner.do(http.MethodPut, "/api/branding/settings", obj{"name": "Example GmbH"}, http.StatusForbidden)

	now := time.Now().UTC()
	key := domain.Sign(domain.Licence{ID: "t", Licensee: "Example GmbH", Issued: now, Until: now.AddDate(1, 0, 0), Seats: 5,
		Features: []string{"branding"}}, priv)
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)

	var view obj
	owner.json(http.MethodPut, "/api/branding/settings", obj{"name": "Example GmbH", "accent": "#FFEE00", "hidePoweredBy": true},
		http.StatusOK, &view)
	if view["applied"] != true || len(view["warnings"].([]any)) != 1 {
		t.Fatalf("settings view: %v", view)
	}
	owner.do(http.MethodPut, "/api/branding/settings", obj{"accent": "yellow"}, http.StatusUnprocessableEntity)

	// Upload: re-encoded as PNG; SVG and junk refused.
	var img bytes.Buffer
	pic := image.NewRGBA(image.Rect(0, 0, 900, 300))
	pic.Set(1, 1, color.RGBA{R: 255, A: 255})
	_ = png.Encode(&img, pic)
	upload := func(kind string, body []byte, want int) {
		req, _ := http.NewRequest(http.MethodPut, srv.URL+"/api/branding/assets/"+kind, bytes.NewReader(body))
		req.Header.Set("X-CSRF-Token", owner.csrf)
		req.Header.Set("Content-Type", "application/octet-stream")
		res, err := owner.http.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		if res.StatusCode != want {
			t.Fatalf("upload %s: %d, want %d", kind, res.StatusCode, want)
		}
	}
	upload("logo", img.Bytes(), http.StatusOK)
	upload("favicon", []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), http.StatusUnprocessableEntity)
	upload("banner", img.Bytes(), http.StatusNotFound)

	b := info()
	if b["name"] != "Example GmbH" || b["custom"] != true || b["hidePoweredBy"] != true || !strings.HasPrefix(b["logo"].(string), "/api/branding/logo?v=") {
		t.Fatalf("branding: %v", b)
	}
	res, err := http.Get(srv.URL + b["logo"].(string))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	logo, err := png.Decode(bytes.NewReader(data))
	if res.StatusCode != http.StatusOK || err != nil || logo.Bounds().Dx() != 512 {
		t.Fatalf("logo: %d %v %v", res.StatusCode, err, logo)
	}
	res, err = http.Get(srv.URL + "/manifest.webmanifest")
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	_ = json.NewDecoder(res.Body).Decode(&manifest)
	_ = res.Body.Close()
	if manifest["name"] != "Example GmbH" {
		t.Fatalf("manifest: %v", manifest)
	}

	// New authenticator entries carry the organisation's name (#305).
	var enrollment obj
	owner.json(http.MethodPost, "/api/auth/totp/setup", nil, http.StatusOK, &enrollment)
	if uri, _ := enrollment["uri"].(string); !strings.Contains(uri, "issuer=Example") {
		t.Fatalf("totp uri: %v", enrollment)
	}

	// Members cannot change it.
	owner.do(http.MethodPost, "/api/auth/members", obj{"name": "ben", "password": "member-password-1"}, http.StatusCreated)
	ben := newClient(t, srv.URL)
	var bs obj
	ben.json(http.MethodPost, "/api/auth/login", obj{"username": "ben", "password": "member-password-1"}, http.StatusOK, &bs)
	ben.csrf = bs["csrfToken"].(string)
	ben.do(http.MethodPut, "/api/branding/settings", obj{"name": "Ben's"}, http.StatusForbidden)

	// Without the key the product shows again; the branding is kept.
	owner.do(http.MethodDelete, "/api/licence", nil, http.StatusNoContent)
	if b := info(); b["name"] != "trisse-home" || b["custom"] != false {
		t.Fatalf("after removing the key: %v", b)
	}
	owner.do(http.MethodPut, "/api/licence", obj{"key": key}, http.StatusOK)
	if b := info(); b["name"] != "Example GmbH" {
		t.Fatalf("after the key is back: %v", b)
	}
}

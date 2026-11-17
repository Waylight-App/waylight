package branding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeSource struct {
	on bool
	b  Branding
}

func (f fakeSource) Branding(context.Context) (Branding, bool) { return f.b, f.on }
func (f fakeSource) Asset(_ context.Context, kind string) ([]byte, string, bool) {
	if _, ok := f.b.Assets[kind]; !ok {
		return nil, "", false
	}
	return []byte("png:" + kind), "image/png", true
}

func get(h http.HandlerFunc, target, kind string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if kind != "" {
		req.SetPathValue("kind", kind)
	}
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestProductBrandingWithoutSource(t *testing.T) {
	var info InfoDTO
	_ = json.Unmarshal(get(InfoHandler(func() Source { return nil }), "/api/branding", "").Body.Bytes(), &info)
	if info.Name != "Waylight" || info.Custom || info.Logo != "" {
		t.Fatalf("%+v", info)
	}
	if rec := get(AssetHandler(func() Source { return nil }), "/api/branding/logo", "logo"); rec.Code != http.StatusNotFound {
		t.Fatalf("asset without source: %d", rec.Code)
	}
}

func TestOrganisationBranding(t *testing.T) {
	at := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	src := fakeSource{on: true, b: Branding{Name: "Example GmbH", Accent: "#0055aa", HidePoweredBy: true, Assets: map[string]time.Time{Logo: at, Favicon: at}}}
	s := func() Source { return src }
	var info InfoDTO
	_ = json.Unmarshal(get(InfoHandler(s), "/api/branding", "").Body.Bytes(), &info)
	if info.Name != "Example GmbH" || !info.Custom || info.Logo != "/api/branding/logo?v=20261116120000" || info.LogoDark != "" || !info.HidePoweredBy {
		t.Fatalf("%+v", info)
	}
	rec := get(AssetHandler(s), "/api/branding/logo?v=1", "logo")
	if rec.Code != http.StatusOK || rec.Body.String() != "png:logo" || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("logo: %d %v", rec.Code, rec.Header())
	}
	if rec := get(AssetHandler(s), "/api/branding/logo-dark", "logo-dark"); rec.Code != http.StatusNotFound {
		t.Fatalf("missing dark logo: %d", rec.Code)
	}
	if rec := get(AssetHandler(s), "/api/branding/../x", "x"); rec.Code != http.StatusNotFound {
		t.Fatalf("unknown kind: %d", rec.Code)
	}
	base := []byte(`{"name":"Waylight","short_name":"Waylight","theme_color":"#3b6ef5","icons":[{"src":"/icons/icon-192.png"}]}`)
	var m map[string]any
	_ = json.Unmarshal(get(ManifestHandler(base, s), "/manifest.webmanifest", "").Body.Bytes(), &m)
	icons, _ := m["icons"].([]any)
	if m["name"] != "Example GmbH" || m["theme_color"] != "#0055aa" || len(icons) != 1 {
		t.Fatalf("manifest %v", m)
	}
	// An expired licence: the stored branding is not applied.
	off := func() Source { return fakeSource{on: false, b: src.b} }
	if body := get(ManifestHandler(base, off), "/manifest.webmanifest", "").Body.String(); body != string(base) {
		t.Fatalf("manifest without licence: %s", body)
	}
	if rec := get(AssetHandler(off), "/api/branding/logo", "logo"); rec.Code != http.StatusNotFound {
		t.Fatalf("logo without licence: %d", rec.Code)
	}
}

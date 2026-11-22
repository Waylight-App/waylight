package controller

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"waylight/internal/modules/setupsignin/domain"
	"waylight/internal/modules/setupsignin/usecase"
)

type fakeFill struct {
	usecase.Pastes
	err   error
	value string
}

func (f *fakeFill) Check(string, string) error { return f.err }
func (f *fakeFill) Fill(_, _, v string) error  { f.value = v; return f.err }

func TestPastePageIsLockedDown(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/paste/abc", nil)
	req.Host = "wl.example"
	req.SetPathValue("code", "abc")
	req.Header.Set("X-Forwarded-Proto", "https")
	NewPasteHandlers(&fakeFill{}).Page(rec, req)
	h, body := rec.Header(), rec.Body.String()
	if rec.Code != 200 || h.Get("Cache-Control") != "no-store" || h.Get("Set-Cookie") != "" {
		t.Fatalf("%d %v", rec.Code, h)
	}
	csp := h.Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'none'") || strings.Contains(csp, "script-src") || strings.Contains(csp, "unsafe") {
		t.Fatalf("csp: %s", csp)
	}
	if !strings.Contains(body, "<style>"+pageStyle+"</style>") || !strings.Contains(csp, "sha256-"+styleHash()) {
		t.Fatal("style does not match the CSP hash")
	}
	if strings.Contains(body, "<script") || strings.Contains(body, "http://") || strings.Contains(body, "src=") {
		t.Fatal("page must not load anything")
	}
	want := "waylight://paste?code=abc&amp;server=https%3A%2F%2Fwl.example"
	if !strings.Contains(body, want) {
		t.Fatalf("app link missing: %s", body)
	}
}

func TestPasteFillPage(t *testing.T) {
	post := func(f *fakeFill, lang string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/paste/abc", strings.NewReader(url.Values{"value": {"s3cret"}}.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept-Language", lang)
		NewPasteHandlers(f).Fill(rec, req)
		return rec
	}
	f := &fakeFill{}
	rec := post(f, "de-DE")
	if rec.Code != 200 || f.value != "s3cret" || !strings.Contains(rec.Body.String(), "Gesendet") || strings.Contains(rec.Body.String(), "s3cret") {
		t.Fatalf("%d %s %q", rec.Code, rec.Body, f.value)
	}
	if rec := post(&fakeFill{err: domain.ErrPasteUnknown}, "en"); rec.Code != 404 || !strings.Contains(rec.Body.String(), "expired") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	if rec := post(&fakeFill{err: domain.ErrPasteValue}, "en"); rec.Code != 400 || !strings.Contains(rec.Body.String(), "<form") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestPasteAppLinkRejectsOddHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/paste/abc", nil)
	req.Host = `x"><script>`
	if appLink(req, "abc") != "" {
		t.Fatal("odd host must not become a link")
	}
}

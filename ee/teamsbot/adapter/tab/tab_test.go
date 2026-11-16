package tab

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTabPage(t *testing.T) {
	h := Handlers{Name: func(context.Context) string { return `Example <GmbH>` }, StartURL: StartURL("https://home.example/"),
		Allowed: func(context.Context) bool { return true }}
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Frame-Options", "DENY") // as the security middleware sets it
	h.Page(rec, httptest.NewRequest(http.MethodGet, "/teams/tab", nil))
	body, csp := rec.Body.String(), rec.Header().Get("Content-Security-Policy")
	if rec.Code != http.StatusOK || rec.Header().Get("X-Frame-Options") != "" {
		t.Fatalf("%d %v", rec.Code, rec.Header())
	}
	for _, want := range []string{"frame-ancestors https://*.cloud.microsoft https://teams.microsoft.com", "script-src 'self' https://res.cdn.office.net",
		"default-src 'none'"} {
		if !strings.Contains(csp, want) {
			t.Errorf("csp lacks %q: %s", want, csp)
		}
	}
	if strings.Contains(csp, "unsafe-inline") {
		t.Error("inline code allowed")
	}
	if !strings.Contains(body, `integrity="`+TeamsJSIntegrity+`"`) || !strings.Contains(body, `href="https://home.example/"`) ||
		!strings.Contains(body, "Example &lt;GmbH&gt;") {
		t.Fatalf("page: %s", body)
	}
}

func TestTabNeedsLicenceAndAddress(t *testing.T) {
	for name, h := range map[string]Handlers{
		"unlicensed": {Name: func(context.Context) string { return "x" }, StartURL: "https://h/", Allowed: func(context.Context) bool { return false }},
		"no address": {Name: func(context.Context) string { return "x" }, Allowed: func(context.Context) bool { return true }},
	} {
		rec := httptest.NewRecorder()
		h.Page(rec, httptest.NewRequest(http.MethodGet, "/teams/tab", nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: %d", name, rec.Code)
		}
	}
}

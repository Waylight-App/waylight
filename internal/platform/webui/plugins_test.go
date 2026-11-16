package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestPluginsHandlerListsTheBuildsPlugins(t *testing.T) {
	for _, tc := range []struct {
		urls []string
		want string
	}{{nil, `{"plugins":[]}`}, {[]string{"/ee/plugin.js"}, `{"plugins":["/ee/plugin.js"]}`}} {
		rec := httptest.NewRecorder()
		PluginsHandler(tc.urls)(rec, httptest.NewRequest(http.MethodGet, "/api/plugins", nil))
		if got := strings.TrimSpace(rec.Body.String()); got != tc.want {
			t.Errorf("got %s, want %s", got, tc.want)
		}
	}
}

func TestPluginFiles(t *testing.T) {
	fsys := fstest.MapFS{
		"plugin.js":          {Data: []byte("export default {}")},
		"assets/chunk-ab.js": {Data: []byte("chunk")},
		".gitkeep":           {Data: nil},
	}
	h := PluginFiles("/ee/", fsys)
	get := func(p string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, p, nil))
		return rec
	}
	rec := get("/ee/plugin.js")
	body, _ := io.ReadAll(rec.Body)
	if rec.Code != http.StatusOK || string(body) != "export default {}" || rec.Header().Get("Cache-Control") != "no-cache" ||
		!strings.Contains(rec.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("entry: %d %q %v", rec.Code, body, rec.Header())
	}
	if rec := get("/ee/assets/chunk-ab.js"); rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("chunk: %d %v", rec.Code, rec.Header())
	}
	for _, p := range []string{"/ee/", "/ee/assets/", "/ee/missing.js", "/ee/.gitkeep", "/ee/../internal/x"} {
		if rec := get(p); rec.Code != http.StatusNotFound {
			t.Errorf("%s: %d", p, rec.Code)
		}
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ee/plugin.js", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST: %d", rec.Code)
	}
}

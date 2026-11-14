package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serve(h http.Handler, path, accept string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if accept != "" {
		req.Header.Set("Accept-Encoding", accept)
	}
	rec := httptest.NewRecorder()
	Compress(h).ServeHTTP(rec, req)
	return rec
}

func TestCompress(t *testing.T) {
	big := `{"points":[` + strings.Repeat(`{"t":"2026-11-14T09:00:00Z","v":12.5},`, 200) + `{}]}`
	jsonHandler := func(body string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Content-Length", "123")
			_, _ = io.WriteString(w, body)
		})
	}

	rec := serve(jsonHandler(big), "/api/system/history", "br, gzip")
	if rec.Header().Get("Content-Encoding") != "gzip" || rec.Header().Get("Content-Length") != "" || !strings.Contains(rec.Header().Get("Vary"), "Accept-Encoding") {
		t.Fatalf("headers %v", rec.Header())
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(zr)
	if string(got) != big || rec.Body.Len() >= len(big)/3 {
		t.Fatalf("round trip: %d bytes compressed to %d", len(big), rec.Body.Len())
	}

	for name, rec := range map[string]*httptest.ResponseRecorder{
		"small answer": serve(jsonHandler(`{"ok":true}`), "/api/x", "gzip"),
		"no gzip":      serve(jsonHandler(big), "/api/x", "gzip;q=0, br"),
		"not the API":  serve(jsonHandler(big), "/index.html", "gzip"),
		"not JSON": serve(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			_, _ = io.WriteString(w, big)
		}), "/api/img", "gzip"),
		"no accept":    serve(jsonHandler(big), "/api/x", ""),
		"error status": serve(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }), "/api/x", "gzip"),
	} {
		if rec.Header().Get("Content-Encoding") != "" {
			t.Errorf("%s: compressed", name)
		}
	}
	if rec := serve(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, big)
	}), "/api/x", "gzip"); rec.Code != http.StatusNotFound || rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("status kept: %d %v", rec.Code, rec.Header())
	}
}

func TestCompressFlush(t *testing.T) {
	rec := serve(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: a\n\n")
		http.NewResponseController(w).Flush()
		_, _ = io.WriteString(w, "data: b\n\n")
	}), "/api/summary/stream", "gzip")
	if rec.Header().Get("Content-Encoding") != "" || rec.Body.String() != "data: a\n\ndata: b\n\n" || !rec.Flushed {
		t.Fatalf("%v %q flushed=%v", rec.Header(), rec.Body.String(), rec.Flushed)
	}
}

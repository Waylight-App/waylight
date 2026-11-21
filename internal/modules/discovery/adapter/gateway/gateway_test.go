package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseRoutes(t *testing.T) {
	table := "Iface\tDestination\tGateway \tFlags\n" +
		"eth0\t0012A8C0\t00000000\t0001\n" +
		"eth0\t00000000\t01B2A8C0\t0003\n"
	a, ok := parseRoutes(strings.NewReader(table))
	if !ok || a.String() != "192.168.178.1" {
		t.Fatalf("%v %v", a, ok)
	}
	if _, ok := parseRoutes(strings.NewReader("Iface\tDestination\tGateway\n")); ok {
		t.Fatal("no default route")
	}
}

func TestProbe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/server/ping" {
			_, _ = w.Write([]byte(`{"res":"pong"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()
	p := NewProber(srv.Client())
	if !p.Probe(context.Background(), srv.URL, "/api/server/ping", "pong") {
		t.Fatal("immich answers")
	}
	if p.Probe(context.Background(), srv.URL, "/status.php", "Nextcloud") {
		t.Fatal("something else answers")
	}
}

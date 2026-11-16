package slackapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIdentityAndOpenSocket(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/auth.test" && r.Header.Get("Authorization") == "Bearer xoxp-1":
			_, _ = w.Write([]byte(`{"ok":true,"team_id":"T1","user_id":"U1"}`))
		case r.URL.Path == "/apps.connections.open" && r.Header.Get("Authorization") == "Bearer xapp-1":
			_, _ = w.Write([]byte(`{"ok":true,"url":"ws://127.0.0.1/link"}`))
		default:
			_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
		}
	}))
	defer srv.Close()
	c := Client{HTTP: srv.Client(), API: srv.URL}
	if team, user, err := c.Identity(context.Background(), "xoxp-1"); err != nil || team != "T1" || user != "U1" {
		t.Fatalf("%s %s %v", team, user, err)
	}
	if _, _, err := c.Identity(context.Background(), "wrong"); err == nil {
		t.Fatal("invalid_auth accepted")
	}
	if u, err := c.OpenSocket(context.Background(), "xapp-1"); err != nil || u != "ws://127.0.0.1/link" {
		t.Fatalf("%s %v", u, err)
	}
	// Over a TLS API a plain ws:// address is refused.
	tls := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"url":"ws://127.0.0.1/link"}`))
	}))
	defer tls.Close()
	if _, err := (Client{HTTP: tls.Client(), API: tls.URL}).OpenSocket(context.Background(), "xapp-1"); err == nil {
		t.Fatal("plain ws:// accepted from a TLS API")
	}
}

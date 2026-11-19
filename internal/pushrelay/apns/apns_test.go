package apns

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/pushrelay"
)

func TestSend(t *testing.T) {
	ek, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	der, _ := x509.MarshalPKCS8PrivateKey(ek)
	bearers := map[string]bool{}
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearers[r.Header.Get("Authorization")] = true
		if r.Header.Get("apns-topic") != "app.getwaylight.mobile" || r.Header.Get("apns-push-type") != "alert" {
			t.Error("headers", r.Header)
		}
		b, _ := io.ReadAll(r.Body)
		lastBody = string(b)
		switch r.URL.Path {
		case "/3/device/gone":
			w.WriteHeader(http.StatusGone)
			_, _ = io.WriteString(w, `{"reason":"Unregistered"}`)
		case "/3/device/bad":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"reason":"BadDeviceToken"}`)
		case "/3/device/busy":
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"reason":"TooManyRequests"}`)
		}
	}))
	defer srv.Close()
	now := time.Unix(1_800_000_000, 0)
	s, err := New(srv.Client(), srv.URL, Config{KeyID: "K", TeamID: "T", Topic: "app.getwaylight.mobile",
		Key: pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := s.Send(ctx, "dev", []byte{1, 2}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(lastBody, `"mutable-content":1`) || !strings.Contains(lastBody, `"p":"AQI="`) {
		t.Fatal(lastBody)
	}
	for _, tok := range []string{"gone", "bad"} {
		if err := s.Send(ctx, tok, []byte{1}); !errors.Is(err, pushrelay.ErrUnregistered) {
			t.Fatal(tok, err)
		}
	}
	if err := s.Send(ctx, "busy", []byte{1}); err == nil || errors.Is(err, pushrelay.ErrUnregistered) {
		t.Fatal(err)
	}
	if len(bearers) != 1 {
		t.Fatal("provider token must be reused:", len(bearers))
	}
	now = now.Add(41 * time.Minute)
	_ = s.Send(ctx, "dev", []byte{1})
	if len(bearers) != 2 {
		t.Fatal("provider token must be renewed")
	}
}

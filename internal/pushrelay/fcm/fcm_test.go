package fcm

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
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
	rk, _ := rsa.GenerateKey(rand.Reader, 2048)
	der, _ := x509.MarshalPKCS8PrivateKey(rk)
	tokens, sends := 0, 0
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokens++
			_ = r.ParseForm()
			if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || strings.Count(r.Form.Get("assertion"), ".") != 2 {
				t.Error("bad grant", r.Form)
			}
			_, _ = io.WriteString(w, `{"access_token":"AT","expires_in":3600}`)
		case "/v1/projects/proj/messages:send":
			sends++
			if r.Header.Get("Authorization") != "Bearer AT" {
				t.Error("auth", r.Header.Get("Authorization"))
			}
			b, _ := io.ReadAll(r.Body)
			lastBody = string(b)
			if strings.Contains(lastBody, `"gone"`) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = io.WriteString(w, `{"error":{"status":"NOT_FOUND","details":[{"errorCode":"UNREGISTERED"}]}}`)
				return
			}
			if strings.Contains(lastBody, `"broken"`) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = io.WriteString(w, `{"name":"x"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	acc, _ := json.Marshal(Account{ProjectID: "proj", ClientEmail: "relay@proj", TokenURI: srv.URL + "/token",
		PrivateKey: string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))})
	now := time.Unix(1_800_000_000, 0)
	s, err := New(srv.Client(), srv.URL, acc, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := s.Send(ctx, "dev", []byte{1, 2}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(lastBody, `"data":{"p":"AQI="}`) || !strings.Contains(lastBody, `"priority":"HIGH"`) {
		t.Fatal(lastBody)
	}
	if err := s.Send(ctx, "gone", []byte{1}); !errors.Is(err, pushrelay.ErrUnregistered) {
		t.Fatal(err)
	}
	if err := s.Send(ctx, "broken", []byte{1}); err == nil || errors.Is(err, pushrelay.ErrUnregistered) {
		t.Fatal(err)
	}
	if tokens != 1 || sends != 3 {
		t.Fatal("access token must be reused:", tokens, sends)
	}
	now = now.Add(56 * time.Minute)
	_ = s.Send(ctx, "dev", []byte{1})
	if tokens != 2 {
		t.Fatal("access token must be renewed before it expires")
	}
	if _, err := New(srv.Client(), srv.URL, []byte(`{}`), time.Now); err == nil {
		t.Fatal("incomplete account must fail")
	}
}

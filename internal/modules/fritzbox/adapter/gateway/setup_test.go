package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"waylight/internal/modules/fritzbox/domain"
	"waylight/internal/shared/kernel"
)

// TestWANWrapsTargetErrors covers the errors the setup assistant explains
// (#618): a refused login and an address that does not speak TR-064.
func TestWANWrapsTargetErrors(t *testing.T) {
	cases := map[string]struct {
		handler http.HandlerFunc
		want    error
	}{
		"refused login": {func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("WWW-Authenticate", `Digest realm="F!Box SOAP-Auth", nonce="n", qop="auth"`)
			w.WriteHeader(http.StatusUnauthorized)
		}, kernel.ErrTargetAuth},
		"html page": {func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html><p>login</html")) }, kernel.ErrTargetWrongService},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()
			_, _, err := NewBox(srv.Client(), time.UTC).WAN(context.Background(), domain.Settings{URL: srv.URL, Username: "home", Password: "secret"})
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

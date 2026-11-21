package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"waylight/internal/modules/adguard/domain"
	"waylight/internal/shared/kernel"
)

// TestStatsWrapTargetErrors covers the errors the setup assistant explains
// (#618) for AdGuard Home and Pi-hole.
func TestStatsWrapTargetErrors(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
		want   error
	}{
		"refused login": {http.StatusUnauthorized, "", kernel.ErrTargetAuth},
		"no api":        {http.StatusNotFound, "", kernel.ErrTargetWrongService},
		"html page":     {http.StatusOK, "<html>", kernel.ErrTargetWrongService},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			ag := domain.Settings{Provider: domain.ProviderAdGuard, URL: srv.URL, Username: "u", Password: "p"}
			if _, err := NewAPI(srv.Client()).Stats(context.Background(), ag); !errors.Is(err, tc.want) {
				t.Fatalf("AdGuard Home: got %v, want %v", err, tc.want)
			}
			pi := domain.Settings{Provider: domain.ProviderPiHole, URL: srv.URL, Password: "p"}
			if _, err := NewPiHole(srv.Client()).Stats(context.Background(), pi); !errors.Is(err, tc.want) {
				t.Fatalf("Pi-hole: got %v, want %v", err, tc.want)
			}
		})
	}
}

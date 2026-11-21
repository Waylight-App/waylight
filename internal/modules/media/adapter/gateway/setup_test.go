package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"waylight/internal/modules/media/domain"
	"waylight/internal/shared/kernel"
)

// TestOverviewWrapsTargetErrors covers the errors the setup assistant
// explains (#618) for both providers.
func TestOverviewWrapsTargetErrors(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
		want   error
	}{
		"refused key": {http.StatusUnauthorized, "", kernel.ErrTargetAuth},
		"forbidden":   {http.StatusForbidden, "", kernel.ErrTargetAuth},
		"no api":      {http.StatusNotFound, "", kernel.ErrTargetWrongService},
		"html page":   {http.StatusOK, "<html>", kernel.ErrTargetWrongService},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			s := domain.Settings{URL: srv.URL, Token: "t"}
			if _, err := NewJellyfin(srv.Client()).Overview(context.Background(), s); !errors.Is(err, tc.want) {
				t.Fatalf("Jellyfin: got %v, want %v", err, tc.want)
			}
			if _, err := NewPlex(srv.Client()).Overview(context.Background(), s); !errors.Is(err, tc.want) {
				t.Fatalf("Plex: got %v, want %v", err, tc.want)
			}
		})
	}
}

package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"waylight/internal/modules/arr/domain"
	"waylight/internal/shared/kernel"
)

// TestCheckWrapsTargetErrors covers the errors the setup assistant explains
// (#618): a refused key, another service and another *arr app.
func TestCheckWrapsTargetErrors(t *testing.T) {
	cases := map[string]struct {
		handler http.HandlerFunc
		want    error
	}{
		"refused key": {func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusUnauthorized) }, kernel.ErrTargetAuth},
		"no api":      {func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }, kernel.ErrTargetWrongService},
		"html page":   {func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) }, kernel.ErrTargetWrongService},
		"radarr":      {func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"appName":"Radarr"}`)) }, kernel.ErrTargetWrongService},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()
			err := NewSonarr(srv.Client()).Check(context.Background(), domain.Instance{App: domain.AppSonarr, URL: srv.URL, APIKey: "k"})
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

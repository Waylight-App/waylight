package gateway

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/shared/kernel"
)

// TestResultsWrapTargetErrors covers the errors the setup assistant explains
// (#618).
func TestResultsWrapTargetErrors(t *testing.T) {
	cases := map[string]struct {
		status int
		body   string
		want   error
	}{
		"refused token": {http.StatusUnauthorized, "", kernel.ErrTargetAuth},
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
			_, err := NewAPI(srv.Client()).Results(context.Background(), domain.Settings{URL: srv.URL, Token: "t"}, time.Now())
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

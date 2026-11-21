package smarthome

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/shared/kernel"
)

// TestRenewer talks to a fake Home Assistant token endpoint: the refresh
// grant carries the client ID, the refresh token is kept (Home Assistant
// sends none) and a revoked one is an authentication failure.
func TestRenewer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.URL.Path != "/auth/token" || r.PostForm.Get("grant_type") != "refresh_token" || r.PostForm.Get("client_id") != "https://home.example/" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.PostForm.Get("refresh_token") != "refresh" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","expires_in":1800,"token_type":"Bearer"}`))
	}))
	defer srv.Close()
	now := time.Unix(1_700_000_000, 0)
	r := newRenewer(srv.Client(), func() time.Time { return now })
	s := domain.Settings{URL: srv.URL, Token: "old", Refresh: "refresh", ClientID: "https://home.example/"}

	got, err := r.Renew(context.Background(), s)
	if err != nil || got.Token != "new-access" || got.Refresh != "refresh" || !got.Expiry.Equal(now.Add(30*time.Minute)) {
		t.Fatalf("renew: %+v %v", got, err)
	}
	s.Refresh = "revoked"
	if _, err := r.Renew(context.Background(), s); !errors.Is(err, kernel.ErrTargetAuth) {
		t.Fatalf("revoked refresh token: %v", err)
	}
}

package setupsignin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

// TestHomeAssistantOAuth: the client ID is Waylight's origin and the
// redirect lies on it (IndieAuth), the code is exchanged at the instance's
// token endpoint, and an answer without refresh token is refused.
func TestHomeAssistantOAuth(t *testing.T) {
	refresh := "refresh"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.URL.Path != "/auth/token" || r.PostForm.Get("grant_type") != "authorization_code" ||
			r.PostForm.Get("client_id") != "https://home.example/" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.PostForm.Get("code") != "good" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_request"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"access","expires_in":1800,"refresh_token":"` + refresh + `","token_type":"Bearer"}`))
	}))
	defer srv.Close()
	now := time.Unix(1_700_000_000, 0)
	public, _ := url.Parse("https://home.example")
	h := newHomeAssistant(srv.Client(), public, func() time.Time { return now })

	if h.ClientID() != "https://home.example/" {
		t.Fatalf("client ID %q", h.ClientID())
	}
	u, _ := url.Parse(h.AuthURL("http://ha.local:8123", "st"))
	q := u.Query()
	if u.Host != "ha.local:8123" || u.Path != "/auth/authorize" || q.Get("client_id") != "https://home.example/" ||
		q.Get("redirect_uri") != "https://home.example"+CallbackPath || q.Get("state") != "st" || q.Get("response_type") != "code" {
		t.Fatalf("authorise URL %s", u)
	}

	g, err := h.Exchange(context.Background(), srv.URL, "good")
	if err != nil || g.Access != "access" || g.Secret != "refresh" || g.ClientID != "https://home.example/" || !g.Expiry.Equal(now.Add(30*time.Minute)) {
		t.Fatalf("exchange: %+v %v", g, err)
	}
	if _, err := h.Exchange(context.Background(), srv.URL, "bad"); !errors.Is(err, kernel.ErrTargetAuth) {
		t.Fatalf("refused code: %v", err)
	}
	refresh = ""
	if _, err := h.Exchange(context.Background(), srv.URL, "good"); !errors.Is(err, kernel.ErrTargetWrongService) {
		t.Fatalf("no refresh token: %v", err)
	}
	if newHomeAssistant(srv.Client(), nil, time.Now).ClientID() != "" {
		t.Fatal("without PUBLIC_URL there is no client ID")
	}
}

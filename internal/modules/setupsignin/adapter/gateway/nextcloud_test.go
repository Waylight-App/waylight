package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"waylight/internal/shared/kernel"
)

// TestLoginFlow runs Login Flow v2 against a fake Nextcloud: the app is
// named Waylight, polls answer 404 until access is granted, and the poll
// goes to the entered address even if Nextcloud names another endpoint.
func TestLoginFlow(t *testing.T) {
	granted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("User-Agent") != "Waylight" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.URL.Path {
		case "/nc/index.php/login/v2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"poll":  map[string]string{"token": "poll-token", "endpoint": "https://elsewhere.example/login/v2/poll"},
				"login": "https://cloud.example/login/v2/flow/xyz",
			})
		case "/nc/index.php/login/v2/poll":
			_ = r.ParseForm()
			if !granted || r.PostForm.Get("token") != "poll-token" {
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"server": "https://cloud.example", "loginName": "anna", "appPassword": "app-pw"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	n := NewNextcloudLogin(srv.Client())
	ctx := context.Background()
	base := srv.URL + "/nc"

	login, token, err := n.Start(ctx, base)
	if err != nil || login != "https://cloud.example/login/v2/flow/xyz" || token != "poll-token" {
		t.Fatalf("start: %q %q %v", login, token, err)
	}
	if _, _, done, err := n.Poll(ctx, base, token); err != nil || done {
		t.Fatalf("pending poll: %v %v", done, err)
	}
	granted = true
	user, password, done, err := n.Poll(ctx, base, token)
	if err != nil || !done || user != "anna" || password != "app-pw" {
		t.Fatalf("granted poll: %q %q %v %v", user, password, done, err)
	}
	if _, _, err := n.Start(ctx, srv.URL+"/other"); !errors.Is(err, kernel.ErrTargetWrongService) {
		t.Fatalf("not Nextcloud: %v", err)
	}
}

func TestLoginFlowWrongAnswers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/html/index.php/login/v2":
			_, _ = w.Write([]byte("<html>router</html>"))
		case "/empty/index.php/login/v2":
			_, _ = w.Write([]byte(`{}`))
		case "/empty/index.php/login/v2/poll":
			_, _ = w.Write([]byte(`{"server":"x"}`))
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	}))
	defer srv.Close()
	n := NewNextcloudLogin(srv.Client())
	ctx := context.Background()
	for _, base := range []string{srv.URL + "/html", srv.URL + "/empty"} {
		if _, _, err := n.Start(ctx, base); !errors.Is(err, kernel.ErrTargetWrongService) {
			t.Errorf("%s: %v", base, err)
		}
	}
	if _, _, _, err := n.Poll(ctx, srv.URL+"/empty", "t"); !errors.Is(err, kernel.ErrTargetWrongService) {
		t.Errorf("poll without app password: %v", err)
	}
	if _, _, err := n.Start(ctx, srv.URL+"/forbidden"); !errors.Is(err, kernel.ErrTargetAuth) {
		t.Errorf("forbidden: %v", err)
	}
}

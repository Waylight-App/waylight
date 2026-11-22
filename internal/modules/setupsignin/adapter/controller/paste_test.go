package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/setupsignin/usecase"
)

type fakePastes struct {
	usecase.Pastes
	ready bool
}

func (f fakePastes) Take(context.Context, string) (string, bool, error) {
	return "key", f.ready, nil
}

// TestPasteTake: the value only appears in the answer that hands it over,
// and no answer is cached.
func TestPasteTake(t *testing.T) {
	for _, tc := range []struct {
		ready bool
		body  string
	}{{false, `{"status":"pending"}`}, {true, `{"status":"ready","value":"key"}`}} {
		rec := httptest.NewRecorder()
		NewPasteHandlers(fakePastes{ready: tc.ready}).Take(rec, httptest.NewRequest(http.MethodGet, "/api/setup/paste/c", nil))
		if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != tc.body || rec.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("ready=%v: %d %v %s", tc.ready, rec.Code, rec.Header(), rec.Body)
		}
	}
}

func (f fakePastes) Create(context.Context) (string, time.Time, error) {
	return "c", time.Unix(1_700_000_300, 0), nil
}

func TestPasteCreate(t *testing.T) {
	rec := httptest.NewRecorder()
	NewPasteHandlers(fakePastes{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/setup/paste", nil))
	if strings.TrimSpace(rec.Body.String()) != `{"code":"c","expiresAt":"2023-11-14T22:18:20Z"}` {
		t.Fatalf("%s", rec.Body)
	}
}

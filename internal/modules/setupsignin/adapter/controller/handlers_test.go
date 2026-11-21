package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"waylight/internal/modules/setupsignin/domain"
	"waylight/internal/modules/setupsignin/usecase"
	"waylight/internal/shared/kernel"
)

type fakeUC struct {
	usecase.SignIns
	callbackErr error
}

func (f fakeUC) Status(context.Context, string) (domain.Status, error) { return domain.Ready, nil }
func (f fakeUC) Callback(context.Context, string, string, string) error {
	return f.callbackErr
}

// TestCallbackPage: the page after the redirect is not cached, sends no
// referrer (the URL carried the code) and says what to do.
func TestCallbackPage(t *testing.T) {
	for _, tc := range []struct {
		err  error
		code int
		text string
	}{{nil, http.StatusOK, "Signed in"}, {kernel.ErrSignInUnknown, http.StatusBadRequest, "Sign-in failed"}} {
		rec := httptest.NewRecorder()
		NewHandlers(fakeUC{callbackErr: tc.err}).Callback(rec, httptest.NewRequest(http.MethodGet, "/api/setup/signin/oauth/callback?state=s&code=c", nil))
		if rec.Code != tc.code || !strings.Contains(rec.Body.String(), tc.text) ||
			rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("Referrer-Policy") != "no-referrer" {
			t.Errorf("%v: %d %v %s", tc.err, rec.Code, rec.Header(), rec.Body)
		}
	}
}

func TestStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).Status(rec, httptest.NewRequest(http.MethodGet, "/api/setup/signin/x", nil))
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"status":"ready"}` {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

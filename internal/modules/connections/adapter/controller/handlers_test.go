package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"waylight/internal/modules/connections/usecase"
)

type fakeUC struct {
	usecase.Connections
	confirmed *bool
}

func (fakeUC) Add(_ context.Context, in usecase.Input) (usecase.View, error) {
	return usecase.View{ID: "1", Service: in.Service, Name: in.Name, URL: in.URL, Fields: []string{"apiKey"}}, nil
}

func (f fakeUC) Remove(_ context.Context, _ string, confirmed bool) error {
	*f.confirmed = confirmed
	return nil
}

// TestAddNeverEchoesTheCredential checks the answer carries field names only.
func TestAddNeverEchoesTheCredential(t *testing.T) {
	rec := httptest.NewRecorder()
	body := `{"service":"immich","name":"Home","url":"http://h","credential":{"apiKey":"very-secret"}}`
	NewHandlers(fakeUC{}).Add(rec, httptest.NewRequest(http.MethodPost, "/api/connections", strings.NewReader(body)))
	if rec.Code != http.StatusCreated || strings.Contains(rec.Body.String(), "very-secret") {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || got["fields"] == nil || got["usedBy"] == nil {
		t.Fatalf("%v %v", got, err)
	}
}

func TestRemoveConfirm(t *testing.T) {
	var confirmed bool
	h := NewHandlers(fakeUC{confirmed: &confirmed})
	for q, want := range map[string]bool{"": false, "?confirm=true": true} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/api/connections/1"+q, nil)
		req.SetPathValue("id", "1")
		h.Remove(rec, req)
		if rec.Code != http.StatusNoContent || confirmed != want {
			t.Errorf("%q: %d %v", q, rec.Code, confirmed)
		}
	}
}

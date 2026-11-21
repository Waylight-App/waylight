package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"waylight/internal/modules/integrations/domain"
	"waylight/internal/modules/integrations/usecase"
	"waylight/internal/shared/kernel"
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
	NewHandlers(fakeUC{}).Add(rec, httptest.NewRequest(http.MethodPost, "/api/integrations/connections", strings.NewReader(body)))
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
		req := httptest.NewRequest(http.MethodDelete, "/api/integrations/connections/1"+q, nil)
		req.SetPathValue("id", "1")
		h.Remove(rec, req)
		if rec.Code != http.StatusNoContent || confirmed != want {
			t.Errorf("%q: %d %v", q, rec.Code, confirmed)
		}
	}
}

func (fakeUC) Check(_ context.Context, id string) error {
	if id == "bad" {
		return kernel.SetupFailure("Immich", kernel.ErrTargetAuth)
	}
	return nil
}

func (fakeUC) Services(context.Context) ([]domain.Spec, error) {
	return []domain.Spec{{Service: "nextcloud", Title: "Nextcloud", Fields: []domain.Field{{Name: "username"}, {Name: "appPassword", Secret: true}}}}, nil
}

func TestCheck(t *testing.T) {
	h := NewHandlers(fakeUC{})
	for id, want := range map[string]int{"good": http.StatusNoContent, "bad": http.StatusUnprocessableEntity} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/integrations/connections/"+id+"/check", nil)
		req.SetPathValue("id", id)
		h.Check(rec, req)
		if rec.Code != want {
			t.Errorf("%s: %d %s", id, rec.Code, rec.Body)
		}
	}
}

func TestServices(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).Services(rec, httptest.NewRequest(http.MethodGet, "/api/integrations/services", nil))
	want := `[{"service":"nextcloud","title":"Nextcloud","fields":[{"name":"username","secret":false,"optional":false},{"name":"appPassword","secret":true,"optional":false}]}]`
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

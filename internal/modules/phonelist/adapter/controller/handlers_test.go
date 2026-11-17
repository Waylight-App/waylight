package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/phonelist/domain"
	"waylight/internal/modules/phonelist/usecase"
)

type fakeUC struct{ usecase.Phonelist }

func (fakeUC) Add(_ context.Context, in domain.EntryInput) (domain.Entry, error) {
	return domain.NewEntry("1", in, time.Unix(0, 0))
}

func (fakeUC) Search(_ context.Context, q string) ([]domain.Entry, error) {
	return []domain.Entry{{ID: "s", Name: q, Phone: "+49 30 1"}}, nil
}

func TestAdd(t *testing.T) {
	for body, want := range map[string]int{
		`{"name":"Anna","phone":"+49 30 1"}`: http.StatusCreated,
		`{"name":"Anna"}`:                    http.StatusUnprocessableEntity,
	} {
		rec := httptest.NewRecorder()
		NewHandlers(fakeUC{}).Add(rec, httptest.NewRequest(http.MethodPost, "/api/phonelist/entries", strings.NewReader(body)))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", body, rec.Code, want)
		}
	}
}

func TestSearchAndDialNumber(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).List(rec, httptest.NewRequest(http.MethodGet, "/api/phonelist/entries?q=anna", nil))
	var got []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got) != 1 || got[0]["dial"] != "+49301" {
		t.Fatalf("%s %v", rec.Body, err)
	}
}

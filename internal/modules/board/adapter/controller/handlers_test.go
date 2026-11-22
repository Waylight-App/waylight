package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/board/domain"
	"waylight/internal/modules/board/usecase"
	"waylight/internal/shared/kernel"
)

type fakeUC struct{ usecase.Board }

func (fakeUC) Create(_ context.Context, in domain.Input, visibility string) (usecase.NoteView, error) {
	day := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	n, err := domain.New("1", in, kernel.Access{Owner: "anna", Visibility: kernel.Visibility(visibility)}, day, day)
	return usecase.NoteView{Note: n, Mine: true}, err
}

func TestCreate(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/board", strings.NewReader(`{"text":"Hi","pinned":true,"expiresOn":"2026-11-30","visibility":"everyone"}`)))
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusCreated || got["pinned"] != true || got["expiresOn"] != "2026-11-30" || got["mine"] != true {
		t.Fatalf("%d %v", rec.Code, got)
	}
	rec = httptest.NewRecorder()
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/board", strings.NewReader(`{"text":""}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty note: %d", rec.Code)
	}
}

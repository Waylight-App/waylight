package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/contracts/domain"
	"waylight/internal/modules/contracts/usecase"
	"waylight/internal/shared/kernel"
)

type fakeUC struct{ usecase.Contracts }

func (fakeUC) Create(_ context.Context, in domain.Input, _ string) (usecase.ContractView, error) {
	now := time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC)
	c, err := domain.New("1", in, kernel.Access{Owner: "anna", Visibility: kernel.Private}, now)
	return usecase.ContractView{Contract: c, Term: c.TermOn(now), Mine: true}, err
}

func TestCreate(t *testing.T) {
	rec := httptest.NewRecorder()
	body := `{"name":"Gym","costCents":3600,"period":"quarterly","termEnd":"2027-03-31","noticeDays":90,"renewMonths":12,"remindDays":14}`
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/contracts", strings.NewReader(body)))
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != http.StatusCreated || got["deadline"] != "2026-12-31" || got["monthlyCents"] != float64(1200) || got["open"] != true {
		t.Fatalf("%d %v", rec.Code, got)
	}
	rec = httptest.NewRecorder()
	NewHandlers(fakeUC{}).Create(rec, httptest.NewRequest(http.MethodPost, "/api/contracts", strings.NewReader(`{"name":"Gym","period":"weekly","termEnd":"2027-03-31"}`)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("weekly: %d", rec.Code)
	}
}

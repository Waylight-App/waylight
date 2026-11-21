package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"waylight/internal/modules/attention/domain"
)

type fakeUC struct{ items []domain.Item }

func (f fakeUC) List(context.Context) ([]domain.Item, error) { return f.items, nil }

func TestList(t *testing.T) {
	since := time.Unix(1_800_000_000, 0)
	h := NewHandlers(fakeUC{items: []domain.Item{{
		ID: "module:immich", Kind: domain.KindConnection, Integration: "immich", WidgetID: "w2", TabID: "t1",
		Title: "immich", Since: since, Action: domain.ActionCheck,
		Reason: domain.Message{Text: "Immich does not answer.", Template: "Immich does not answer."},
	}}})
	rec := httptest.NewRecorder()
	h.List(rec, httptest.NewRequest(http.MethodGet, "/api/attention", nil))
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Code)
	}
	var got ListDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Count != 1 || got.Items[0].Reason.MessageKey == "" || got.Items[0].Reason.Message != "Immich does not answer." ||
		got.Items[0].Kind != "connection" || got.Items[0].Action != "check" || got.Items[0].WidgetID != "w2" {
		t.Fatalf("%+v", got)
	}
	rec = httptest.NewRecorder()
	NewHandlers(fakeUC{items: []domain.Item{}}).List(rec, httptest.NewRequest(http.MethodGet, "/api/attention", nil))
	if rec.Body.String() != "{\"items\":[],\"count\":0}\n" {
		t.Fatalf("empty list: %q", rec.Body.String())
	}
}

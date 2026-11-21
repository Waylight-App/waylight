package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"waylight/internal/modules/attention/adapter/presenter"
	"waylight/internal/modules/attention/domain"
)

type fakeUC struct {
	items []domain.Item
	fresh *bool
}

func (f fakeUC) List(_ context.Context, fresh bool) ([]domain.Item, error) {
	if f.fresh != nil {
		*f.fresh = fresh
	}
	return f.items, nil
}

func (f fakeUC) Connections(context.Context, bool) ([]domain.Connection, error) {
	return []domain.Connection{{Integration: "immich", Instance: true, Status: domain.Status{State: domain.StateOK},
		Widgets: []domain.WidgetRef{{ID: "w1", TabID: "t1", Kind: "immich"}}}}, nil
}

func TestConnections(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(fakeUC{}).Connections(rec, httptest.NewRequest(http.MethodGet, "/api/attention/connections", nil))
	var got []presenter.ConnectionDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || len(got) != 1 || got[0].State != "ok" ||
		got[0].Reason != nil || len(got[0].Widgets) != 1 {
		t.Fatalf("%s %v", rec.Body.String(), err)
	}
}

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
	var got presenter.ListDTO
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
	var fresh bool
	NewHandlers(fakeUC{fresh: &fresh}).List(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/attention?refresh=true", nil))
	if !fresh {
		t.Fatal("refresh=true asks for a fresh list")
	}
}

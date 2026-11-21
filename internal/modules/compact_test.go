package modules

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

type compactFake struct {
	state kernel.CompactState
	err   error
}

func (f compactFake) CompactState(context.Context) (kernel.CompactState, error) {
	return f.state, f.err
}

func TestCompactHandler(t *testing.T) {
	at := time.Unix(1_800_000_000, 0).UTC()
	h := CompactHandler(compactFake{state: kernel.CompactState{
		Title:     kernel.NewTextf("{count} open", "count", 3),
		Detail:    kernel.NewTextf("{count} overdue", "count", 1),
		Level:     kernel.CompactAttention,
		Count:     kernel.IntPtr(3),
		UpdatedAt: at,
		Action: &kernel.CompactAction{ID: "presence.set", Label: kernel.NewText("My status"), Method: http.MethodPut, Path: "/api/presence/me",
			Choices: []kernel.CompactChoice{{ID: "office", Label: kernel.NewText("Office"), Body: map[string]any{"status": "office"}, Selected: true}}},
	}})
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/api/todos/compact", nil))
	var got CompactResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || rec.Code != http.StatusOK {
		t.Fatal(rec.Code, err)
	}
	if got.Title.Text != "3 open" || got.Title.Key != kernel.MessageKey("{count} open") || got.Title.Params["count"] != "3" {
		t.Fatalf("title %+v", got.Title)
	}
	if got.Detail == nil || got.Detail.Text != "1 overdue" || got.Level != "attention" || *got.Count != 3 || !got.UpdatedAt.Equal(at) {
		t.Fatalf("%+v", got)
	}
	if got.Action == nil || len(got.Action.Choices) != 1 || !got.Action.Choices[0].Selected {
		t.Fatalf("action %+v", got.Action)
	}

	// No detail, no level: none.
	empty := NewCompactResponse(kernel.CompactState{Title: kernel.NewText("Nothing to do")})
	if empty.Level != "none" || empty.Detail != nil || empty.Action != nil {
		t.Fatalf("%+v", empty)
	}

	rec = httptest.NewRecorder()
	CompactHandler(compactFake{err: kernel.ErrNotFound})(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatal(rec.Code)
	}
	if CompactRoute("todos") != "GET /api/todos/compact" {
		t.Fatal(CompactRoute("todos"))
	}
}

type paramFake struct{}

func (paramFake) CompactStateFor(_ context.Context, p url.Values) (kernel.CompactState, error) {
	if p.Get("entity") == "" {
		return kernel.CompactState{}, kernel.NewValidationError("entity", "missing")
	}
	return kernel.CompactState{Title: kernel.NewTextf("{name}", "name", p.Get("entity"))}, nil
}

func (paramFake) CompactParams() []string { return []string{"entity"} }

func TestCompactParamHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	CompactParamHandler(paramFake{})(rec, httptest.NewRequest(http.MethodGet, "/api/x/compact?entity=switch.coffee", nil))
	var got CompactResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil || rec.Code != http.StatusOK || got.Title.Text != "switch.coffee" {
		t.Fatalf("%d %+v %v", rec.Code, got, err)
	}
	rec = httptest.NewRecorder()
	CompactParamHandler(paramFake{})(rec, httptest.NewRequest(http.MethodGet, "/api/x/compact", nil))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing parameter answered %d", rec.Code)
	}
}

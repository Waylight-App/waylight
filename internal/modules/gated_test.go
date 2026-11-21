package modules

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"slices"
	"testing"

	"waylight/internal/shared/kernel"
)

// muxRouter registers both kinds of routes on one ServeMux, without a session check.
type muxRouter struct{ mux *http.ServeMux }

func (m muxRouter) Public(p string, h http.HandlerFunc)    { m.mux.HandleFunc(p, h) }
func (m muxRouter) Protected(p string, h http.HandlerFunc) { m.mux.HandleFunc(p, h) }

// gatedModule is a module whose access the test sets.
type gatedModule struct {
	Base
	name   string
	access Access
}

func (g *gatedModule) Name() string                  { return g.name }
func (g *gatedModule) DataKind() DataKind            { return InstanceData }
func (g *gatedModule) WidgetKinds() []string         { return []string{g.name + "-widget"} }
func (g *gatedModule) Access(context.Context) Access { return g.access }
func (g *gatedModule) RegisterRoutes(r Router) {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }
	r.Protected("GET /api/"+g.name, ok)
	r.Protected("PUT /api/"+g.name, ok)
}

type plainModule struct{ Base }

func (plainModule) Name() string       { return "notes" }
func (plainModule) DataKind() DataKind { return PersonalContent }

func TestGatedRouterAppliesAccessPerRequest(t *testing.T) {
	m := &gatedModule{name: "branding"}
	mux := http.NewServeMux()
	m.RegisterRoutes(GatedRouter(muxRouter{mux}, m))
	status := func(method string) int {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, "/api/branding", nil))
		return rec.Code
	}
	for _, tc := range []struct {
		access   Access
		get, put int
	}{
		{AccessNone, http.StatusForbidden, http.StatusForbidden},
		{AccessRead, http.StatusNoContent, http.StatusForbidden},
		{AccessFull, http.StatusNoContent, http.StatusNoContent},
	} {
		// The same routes, so a key entered or expiring takes effect without a restart.
		m.access = tc.access
		if got := status(http.MethodGet); got != tc.get {
			t.Errorf("access %d: GET = %d, want %d", tc.access, got, tc.get)
		}
		if got := status(http.MethodPut); got != tc.put {
			t.Errorf("access %d: PUT = %d, want %d", tc.access, got, tc.put)
		}
	}
}

func TestModulesListLeavesOutGatedModulesWithoutAccess(t *testing.T) {
	hidden := &gatedModule{name: "branding", access: AccessNone}
	expired := &gatedModule{name: "announcements", access: AccessRead}
	r := &Registry{}
	r.Register(plainModule{})
	r.Register(hidden)
	r.Register(expired)

	rec := httptest.NewRecorder()
	r.Handler([]string{"clock"})(rec, httptest.NewRequest(http.MethodGet, "/api/modules", nil))
	var got InfoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Modules, []string{"notes", "announcements"}) {
		t.Errorf("modules = %v", got.Modules)
	}
	if slices.Contains(got.WidgetKinds, "branding-widget") || !slices.Contains(got.WidgetKinds, "announcements-widget") {
		t.Errorf("widget kinds = %v", got.WidgetKinds)
	}
	if !slices.Equal(got.ReadOnly, []string{"announcements"}) {
		t.Errorf("read only = %v", got.ReadOnly)
	}
	if _, ok := got.DataKinds["branding"]; ok {
		t.Errorf("data kinds list a hidden module: %v", got.DataKinds)
	}
}

type compactModule struct{ Base }

func (compactModule) Name() string       { return "todos" }
func (compactModule) DataKind() DataKind { return PersonalContent }
func (compactModule) CompactState(context.Context) (kernel.CompactState, error) {
	return kernel.CompactState{}, nil
}

type paramCompactModule struct{ Base }

func (paramCompactModule) Name() string       { return "smarthome" }
func (paramCompactModule) DataKind() DataKind { return InstanceData }
func (paramCompactModule) CompactStateFor(context.Context, url.Values) (kernel.CompactState, error) {
	return kernel.CompactState{}, nil
}
func (paramCompactModule) CompactParams() []string { return []string{"entity"} }

func TestModulesListNamesCompactStates(t *testing.T) {
	r := &Registry{}
	r.Register(plainModule{})
	r.Register(compactModule{})
	r.Register(paramCompactModule{})
	r.Register(&gatedModule{name: "branding", access: AccessNone})

	rec := httptest.NewRecorder()
	r.Handler(nil)(rec, httptest.NewRequest(http.MethodGet, "/api/modules", nil))
	var got InfoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := []CompactInfo{
		{Module: "todos", Path: "/api/todos/compact", Params: []string{}},
		{Module: "smarthome", Path: "/api/smarthome/compact", Params: []string{"entity"}},
	}
	if !reflect.DeepEqual(got.Compact, want) {
		t.Errorf("compact = %+v", got.Compact)
	}
}

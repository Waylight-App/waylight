package modules

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"trisse-home/internal/shared/kernel"
)

type sinkFunc func(Change)

func (f sinkFunc) RecordChange(_ context.Context, c Change) { f(c) }

// kindModule is a module of a given data kind with a few routes.
type kindModule struct {
	Base
	name string
	kind DataKind
}

func (k kindModule) Name() string       { return k.name }
func (k kindModule) DataKind() DataKind { return k.kind }
func (k kindModule) RegisterRoutes(r Router) {
	ok := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }
	r.Protected("GET /api/"+k.name+"/{id}", ok)
	r.Protected("PUT /api/"+k.name+"/{id}", ok)
	r.Protected("PUT /api/"+k.name+"/{id}/visibility", ok)
	r.Protected("DELETE /api/"+k.name+"/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) })
}

func TestChangeRouterRecordsChangesToInstanceDataAndSharing(t *testing.T) {
	var got []Change
	sink := sinkFunc(func(c Change) { got = append(got, c) })
	mux := http.NewServeMux()
	for _, m := range []Module{kindModule{name: "services", kind: InstanceData}, kindModule{name: "notes", kind: PersonalContent}} {
		m.RegisterRoutes(ChangeRouter(muxRouter{mux}, m, sink))
	}
	do := func(method, path string) {
		req := httptest.NewRequest(method, path, nil)
		req = req.WithContext(kernel.WithPrincipal(req.Context(), kernel.Principal{AccountID: "m1", Role: "member"}))
		mux.ServeHTTP(httptest.NewRecorder(), req)
	}
	do(http.MethodGet, "/api/services/s1")         // reads are not changes
	do(http.MethodPut, "/api/services/s1")         // instance data: recorded
	do(http.MethodDelete, "/api/services/s1")      // failed: not recorded
	do(http.MethodPut, "/api/notes/n1")            // personal content: not recorded
	do(http.MethodPut, "/api/notes/n1/visibility") // sharing: recorded
	if len(got) != 2 {
		t.Fatalf("recorded %d changes: %+v", len(got), got)
	}
	want := Change{AccountID: "m1", Role: "member", Module: "services", Route: "PUT /api/services/{id}", ObjectID: "s1", Status: http.StatusNoContent}
	if got[0] != want {
		t.Errorf("first change %+v", got[0])
	}
	if got[1].Module != "notes" || got[1].Route != "PUT /api/notes/{id}/visibility" || got[1].ObjectID != "n1" {
		t.Errorf("sharing change %+v", got[1])
	}
}

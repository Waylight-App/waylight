package modules

import (
	"context"
	"net/http"
	"strings"

	"waylight/internal/shared/kernel"
)

// Change is one successful change made through a module's API: who, what
// kind of change, on which object. It never holds the content itself.
type Change struct {
	AccountID string
	Role      string
	Module    string
	// Route is the pattern as registered, e.g. "PUT /api/notes/{id}".
	Route string
	// ObjectID is the {id} of the route, if it has one.
	ObjectID string
	Status   int
}

// ChangeSink receives changes (output port). Only the ee/ audit log (#303)
// sets one; without it, ChangeRouter is not used at all.
type ChangeSink interface {
	RecordChange(ctx context.Context, c Change)
}

// recordable reports whether changes to a module are logged: instance data
// and sharing (visibility) are, personal content and settings of an account
// are not, so a member's private notes never show up in an admin's log.
func recordable(kind DataKind, pattern string) bool {
	return kind == InstanceData || strings.HasSuffix(pattern, "/visibility")
}

// ChangeRouter registers routes like the wrapped router and reports every
// successful state-changing request of a protected route to the sink.
func ChangeRouter(r Router, m Module, sink ChangeSink) Router {
	return changeRouter{Router: r, m: m, sink: sink}
}

type changeRouter struct {
	Router
	m    Module
	sink ChangeSink
}

// statusRecorder remembers the status a handler wrote.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	return s.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach the original writer.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

func (cr changeRouter) Protected(pattern string, h http.HandlerFunc) {
	method, _, _ := strings.Cut(pattern, " ")
	if method == http.MethodGet || method == http.MethodHead || !recordable(cr.m.DataKind(), pattern) {
		cr.Router.Protected(pattern, h)
		return
	}
	cr.Router.Protected(pattern, func(w http.ResponseWriter, req *http.Request) {
		rec := &statusRecorder{ResponseWriter: w}
		h(rec, req)
		if rec.status == 0 {
			rec.status = http.StatusOK
		}
		if rec.status >= 400 {
			return
		}
		p := kernel.PrincipalFrom(req.Context())
		cr.sink.RecordChange(req.Context(), Change{AccountID: p.AccountID, Role: p.Role, Module: cr.m.Name(), Route: pattern,
			ObjectID: req.PathValue("id"), Status: rec.status})
	})
}

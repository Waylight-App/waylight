package modules

import (
	"context"
	"net/http"

	"trisse-home/internal/shared/httpapi"
	"trisse-home/internal/shared/kernel"
)

// Access is how far a module may be used right now (#307).
type Access int

const (
	// AccessNone hides the module: it is left out of /api/modules and its
	// routes answer 403.
	AccessNone Access = iota
	// AccessRead keeps the data readable but refuses changes, for example
	// while a licence key has expired.
	AccessRead
	// AccessFull is normal use.
	AccessFull
)

// Gated is implemented by modules that are only available under a condition
// the core does not know about, such as the features in ee/ that need a
// licence key (ADR 0017, ADR 0020). The core never imports ee/; it only asks
// this interface, so an ee/ module cannot forget the check: the router and
// /api/modules apply it to every gated module.
type Gated interface {
	Access(ctx context.Context) Access
}

// ErrNotAvailable answers a request to a gated module without access.
var ErrNotAvailable = kernel.NewPublicError(kernel.ErrForbidden, "This feature needs a valid licence key.")

// ErrReadOnly answers a change to a gated module with read access only.
var ErrReadOnly = kernel.NewPublicError(kernel.ErrForbidden, "The licence key has expired or covers fewer accounts than this instance has; the data stays readable, but changes need a valid key.")

// GatedRouter registers routes like the wrapped router, but checks the
// module's access on every request: none refuses everything, read refuses
// everything but GET and HEAD.
func GatedRouter(r Router, g Gated) Router { return gatedRouter{r, g} }

type gatedRouter struct {
	Router
	g Gated
}

func (gr gatedRouter) guard(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		switch gr.g.Access(req.Context()) {
		case AccessFull:
		case AccessRead:
			if req.Method != http.MethodGet && req.Method != http.MethodHead {
				httpapi.WriteDomainError(w, req, ErrReadOnly)
				return
			}
		default:
			httpapi.WriteDomainError(w, req, ErrNotAvailable)
			return
		}
		h(w, req)
	}
}

func (gr gatedRouter) Public(pattern string, h http.HandlerFunc) {
	gr.Router.Public(pattern, gr.guard(h))
}

func (gr gatedRouter) Protected(pattern string, h http.HandlerFunc) {
	gr.Router.Protected(pattern, gr.guard(h))
}

// accessOf is full for modules that are not gated.
func accessOf(ctx context.Context, m Module) Access {
	if g, ok := m.(Gated); ok {
		return g.Access(ctx)
	}
	return AccessFull
}

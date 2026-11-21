package modules

import (
	"context"
	"net/http"

	"waylight/internal/shared/httpapi"
	"waylight/internal/shared/kernel"
)

// DocsBaseURL is the published documentation; help entries link pages
// below it.
const DocsBaseURL = "https://getwaylight.app/docs/"

// WidgetHelp is the short in-app help of one widget kind (#678): what the
// widget shows and how to set it up, in a sentence or two each. The server
// owns the texts so the web interface and the app say the same.
type WidgetHelp struct {
	Kind  string
	Shows kernel.Text
	SetUp kernel.Text
	// Docs is the page relative to DocsBaseURL, optionally with an anchor,
	// e.g. "modules/immich/" or "host-integrations/#wireguard".
	Docs string
}

// Helpful is implemented by modules that back widget kinds. Every widget
// kind has a help entry; cmd/server tests that none is missing.
type Helpful interface {
	WidgetHelp() []WidgetHelp
}

// HelpOf returns the help entries of m, or nil when it has none.
func HelpOf(m Module) []WidgetHelp {
	if h, ok := m.(Helpful); ok {
		return h.WidgetHelp()
	}
	return nil
}

// HelpHandler lists the help of every widget kind the caller can see, the
// same kinds GET /api/modules names, plus the frontend-only ones in static.
func (r *Registry) HelpHandler(static []WidgetHelp) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		res := HelpResponse{Widgets: []WidgetHelpInfo{}}
		add := func(entries []WidgetHelp) {
			for _, e := range entries {
				res.Widgets = append(res.Widgets, WidgetHelpInfo{
					Kind:    e.Kind,
					Shows:   compactText(e.Shows),
					SetUp:   compactText(e.SetUp),
					DocsURL: DocsBaseURL + e.Docs,
				})
			}
		}
		add(static)
		for _, m := range r.visible(req.Context()) {
			add(HelpOf(m))
		}
		httpapi.WriteJSON(w, http.StatusOK, res)
	}
}

// visible returns the modules the caller may see: gated modules without
// access and modules an admin hid from members are left out.
func (r *Registry) visible(ctx context.Context) []Module {
	member := kernel.PrincipalFrom(ctx).Role == "member"
	var out []Module
	for _, m := range r.modules {
		if accessOf(ctx, m) == AccessNone || (member && r.memberMayUse != nil && !r.memberMayUse(ctx, m.Name())) {
			continue
		}
		out = append(out, m)
	}
	return out
}

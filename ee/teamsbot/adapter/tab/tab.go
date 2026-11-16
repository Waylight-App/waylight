// Package tab serves the page of the Microsoft Teams tab (#334, ADR 0023):
// the instance's name and a button that opens the start page in the
// browser. Signing in happens there, never inside Teams, so the session
// cookie stays strict (decision on #334).
//
// Ring: Interface adapters (controller).
package tab

import (
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"strings"
)

// TeamsJS is the Microsoft Teams client library the page must load and
// initialise, or Teams does not show it. It is pinned and checked with
// Subresource Integrity, so a changed file at the CDN is not run.
const (
	TeamsJSURL       = "https://res.cdn.office.net/teams-js/2.58.0/js/MicrosoftTeams.min.js"
	TeamsJSIntegrity = "sha384-sMo/EOqQPa1OCup2/VBRum7oFeOS22kvs8WGDc45xK53tUVZzFR5WwFYDUYN1LO4"
)

// frameAncestors are the hosts that may show the page, from Microsoft's
// "Requirements for building tabs": the new *.cloud.microsoft hosts, Teams,
// the Microsoft 365 app and Outlook.
const frameAncestors = "https://*.cloud.microsoft https://teams.microsoft.com https://*.teams.microsoft.com " +
	"https://*.microsoft365.com https://*.office.com https://outlook.office.com https://outlook.office365.com"

// policy replaces the app's Content-Security-Policy for this page only:
// framed by Microsoft's hosts, scripts from this instance and the pinned
// library, no inline code.
const policy = "default-src 'none'; script-src 'self' https://res.cdn.office.net; style-src 'self'; img-src 'self'; " +
	"connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors " + frameAncestors

//go:embed tab.html
var pageSource string

//go:embed tab.js
var script []byte

//go:embed tab.css
var style []byte

var page = template.Must(template.New("tab").Parse(pageSource))

// Handlers serve the tab.
type Handlers struct {
	// Name is the instance's name (branding, #305).
	Name func(ctx context.Context) string
	// StartURL is PUBLIC_URL; the button opens it.
	StartURL string
	// Allowed tells whether the licence covers chat apps.
	Allowed func(ctx context.Context) bool
}

func (h Handlers) framed(w http.ResponseWriter, contentType string) {
	hd := w.Header()
	hd.Set("Content-Security-Policy", policy)
	hd.Del("X-Frame-Options")
	hd.Set("Content-Type", contentType)
	hd.Set("Cache-Control", "no-cache")
}

// Page handles GET /teams/tab.
func (h Handlers) Page(w http.ResponseWriter, r *http.Request) {
	if !h.Allowed(r.Context()) || h.StartURL == "" {
		http.NotFound(w, r)
		return
	}
	h.framed(w, "text/html; charset=utf-8")
	_ = page.Execute(w, struct {
		Name, StartURL, TeamsJS, Integrity string
	}{h.Name(r.Context()), h.StartURL, TeamsJSURL, TeamsJSIntegrity})
}

// Script handles GET /teams/tab.js.
func (h Handlers) Script(w http.ResponseWriter, r *http.Request) {
	if !h.Allowed(r.Context()) {
		http.NotFound(w, r)
		return
	}
	h.framed(w, "text/javascript; charset=utf-8")
	_, _ = w.Write(script)
}

// Style handles GET /teams/tab.css.
func (h Handlers) Style(w http.ResponseWriter, r *http.Request) {
	if !h.Allowed(r.Context()) {
		http.NotFound(w, r)
		return
	}
	h.framed(w, "text/css; charset=utf-8")
	_, _ = w.Write(style)
}

// StartURL turns PUBLIC_URL into the address the button opens.
func StartURL(publicURL string) string { return strings.TrimSuffix(publicURL, "/") + "/" }

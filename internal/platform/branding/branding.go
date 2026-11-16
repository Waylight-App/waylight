// Package branding serves how the instance presents itself: its name,
// logos, favicon and brand colours on the sign-in page, in the header and
// in the PWA manifest (#305). The core always shows the product; the ee/
// branding module supplies a Source when an organisation set its own and the
// licence covers it, so the core never depends on ee/.
package branding

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Asset kinds.
const (
	Logo     = "logo"
	LogoDark = "logo-dark"
	Favicon  = "favicon"
)

// Branding is what the interface shows.
type Branding struct {
	Name string
	// Accent and Secondary are "#rrggbb" or empty for the theme's own.
	Accent, Secondary string
	HidePoweredBy     bool
	// Assets lists the kinds an organisation uploaded; the rest are the
	// product's.
	Assets map[string]time.Time
}

// Product is the branding without an organisation's own.
var Product = Branding{Name: "trisse-home"}

// Source supplies an organisation's branding (output port, implemented by
// ee/branding). ok is false when none applies, for example without a valid
// licence key; the product branding is shown then.
type Source interface {
	Branding(ctx context.Context) (b Branding, ok bool)
	Asset(ctx context.Context, kind string) (data []byte, contentType string, ok bool)
}

// Current returns the branding to show; src may be nil.
func Current(ctx context.Context, src Source) (Branding, bool) {
	if src != nil {
		if b, ok := src.Branding(ctx); ok {
			if b.Name == "" {
				b.Name = Product.Name
			}
			return b, true
		}
	}
	return Product, false
}

// InfoDTO is GET /api/branding.
type InfoDTO struct {
	Name          string `json:"name"`
	Custom        bool   `json:"custom"`
	Accent        string `json:"accent,omitempty"`
	Secondary     string `json:"secondary,omitempty"`
	HidePoweredBy bool   `json:"hidePoweredBy"`
	// Logo, LogoDark and Favicon are addresses, empty for the product's.
	Logo     string `json:"logo,omitempty"`
	LogoDark string `json:"logoDark,omitempty"`
	Favicon  string `json:"favicon,omitempty"`
}

func assetURL(b Branding, kind string) string {
	at, ok := b.Assets[kind]
	if !ok {
		return ""
	}
	// The version in the query changes with every upload, so the address
	// can be cached for good.
	return "/api/branding/" + kind + "?v=" + at.UTC().Format("20060102150405")
}

// InfoHandler answers GET /api/branding. It is public: the sign-in page
// shows the branding before anyone signs in.
func InfoHandler(src func() Source) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, custom := Current(r.Context(), src())
		out := InfoDTO{Name: b.Name, Custom: custom, Accent: b.Accent, Secondary: b.Secondary, HidePoweredBy: b.HidePoweredBy,
			Logo: assetURL(b, Logo), LogoDark: assetURL(b, LogoDark), Favicon: assetURL(b, Favicon)}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		_ = json.NewEncoder(w).Encode(out)
	}
}

// AssetHandler answers GET /api/branding/{kind} with an uploaded image; 404
// when the organisation has none, so the client keeps the product's.
func AssetHandler(src func() Source) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := r.PathValue("kind")
		s := src()
		if s == nil || (kind != Logo && kind != LogoDark && kind != Favicon) {
			http.NotFound(w, r)
			return
		}
		if _, ok := s.Branding(r.Context()); !ok {
			http.NotFound(w, r)
			return
		}
		data, ct, ok := s.Asset(r.Context(), kind)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.URL.Query().Get("v") != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		_, _ = w.Write(data)
	}
}

// ManifestHandler answers GET /manifest.webmanifest: the app's manifest with
// the organisation's name and icon when one applies. base is the built
// manifest; without branding it is served as built.
func ManifestHandler(base []byte, src func() Source) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json")
		w.Header().Set("Cache-Control", "no-cache")
		b, custom := Current(r.Context(), src())
		if !custom {
			_, _ = w.Write(base)
			return
		}
		var m map[string]any
		if err := json.Unmarshal(base, &m); err != nil || m == nil {
			// A build without the frontend has no manifest to start from.
			m = map[string]any{"start_url": "/", "scope": "/", "display": "standalone"}
		}
		m["name"], m["short_name"] = b.Name, b.Name
		if b.Accent != "" {
			m["theme_color"] = b.Accent
		}
		if u := assetURL(b, Favicon); u != "" {
			m["icons"] = []map[string]string{{"src": u, "sizes": "any", "type": "image/png", "purpose": "any"}}
		}
		_ = json.NewEncoder(w).Encode(m)
	}
}

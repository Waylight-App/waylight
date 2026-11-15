package kernel

import (
	"net/url"
	"regexp"
	"strings"
)

// Background values (#277), shared by the global background (appearance)
// and the per-tab background (dashboard):
//
//   - "" uses the default (or, for a tab, the global background)
//   - an http(s) URL of an image
//   - "/api/appearance/backgrounds/<id>", a picture uploaded to this server
//   - "builtin:<name>", a gradient or colour that ships with the frontend
const (
	// UploadedBackgroundPrefix is the path of uploaded backgrounds.
	UploadedBackgroundPrefix = "/api/appearance/backgrounds/"
	// BuiltinBackgroundPrefix marks a built-in background.
	BuiltinBackgroundPrefix = "builtin:"
	// MaxBackgroundLength limits stored background values.
	MaxBackgroundLength = 2048
)

var (
	uploadedBackground = regexp.MustCompile(`^/api/appearance/backgrounds/[A-Za-z0-9_-]{1,64}$`)
	builtinBackground  = regexp.MustCompile(`^builtin:[a-z0-9-]{1,32}$`)
)

// ParseBackground validates a background value and returns it trimmed.
// Unknown built-in names are accepted; the frontend falls back to the
// default for names it does not know, so an older client never breaks.
func ParseBackground(field, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	switch {
	case raw == "", uploadedBackground.MatchString(raw), builtinBackground.MatchString(raw):
		return raw, nil
	case isImageURL(raw):
		return raw, nil
	}
	return "", NewValidationError(field, "The background must be an http(s) image URL, an uploaded picture or a built-in background.")
}

func isImageURL(raw string) bool {
	// Quotes, parentheses and whitespace could break out of url(...) in CSS.
	if len(raw) > MaxBackgroundLength || strings.ContainsAny(raw, "\"'()\\ \n\t") {
		return false
	}
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

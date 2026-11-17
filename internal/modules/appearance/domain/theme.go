package domain

import (
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Theme gallery (#111): a theme sets design tokens only, never layout or
// free CSS, so app updates cannot break it and an imported theme cannot load
// anything or hide parts of the page.

// ThemeTokens are the design tokens a theme may set (see web tokens.css).
var ThemeTokens = []string{
	"--color-bg", "--color-surface", "--color-surface-2", "--color-border", "--color-text", "--color-text-muted",
	"--color-accent", "--color-accent-contrast", "--color-accent-text", "--color-danger", "--color-success", "--color-warning",
	"--radius", "--radius-small",
	"--font-sans", "--font-display", "--shadow", "--shadow-pop", "--glow", "--page-gradient",
}

// FontStacks are the typefaces a theme can pick for --font-sans and
// --font-display, by keyword (#505). A theme never names a font file or URL:
// Figtree and Fraunces ship with the app (web/src/shared/theme/fonts.css),
// the others are installed on every system.
var FontStacks = map[string]string{
	"system":   "system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', sans-serif",
	"figtree":  "'Figtree Variable', Figtree, system-ui, sans-serif",
	"fraunces": "'Fraunces Variable', Fraunces, Georgia, serif",
	"serif":    "Georgia, 'Times New Roman', serif",
}

// requiredTokens make a theme readable.
var requiredTokens = []string{"--color-bg", "--color-surface", "--color-text", "--color-accent"}

const (
	hexAlpha     = `#(?:[0-9a-f]{6}|[0-9a-f]{8})`
	shadowLength = `(?:0|-?[0-9]{1,2}px)`
	shadowLayer  = shadowLength + ` ` + shadowLength + ` (?:0|[0-9]{1,3}px)(?: ` + shadowLength + `)? ` + hexAlpha
)

var (
	themeID    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,39}$`)
	colorValue = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	radiusVal  = regexp.MustCompile(`^(?:[0-9]|[1-2][0-9]|3[0-2])px$`)
	// Shadows: "none" or up to three layers of offsets, blur, optional
	// spread and a hex colour, optionally with alpha (#rrggbbaa).
	shadowVal = regexp.MustCompile(`^(?:none|` + shadowLayer + `(?:, ` + shadowLayer + `){0,2})$`)
	// Glow: a soft light around a surface, only blur radius and colour.
	glowVal = regexp.MustCompile(`^(?:none|0 0 [0-9]{1,3}px ` + hexAlpha + `)$`)
	// Page gradient: two colours from top to bottom.
	gradientVal = regexp.MustCompile(`^(?:none|linear-gradient\(180deg, ` + hexAlpha + `, ` + hexAlpha + `\))$`)
)

// validToken reports whether v is an allowed value for token k. Every kind
// has a fixed syntax, so a theme cannot smuggle in url(), expressions or
// anything else that is not a colour, a length, a font keyword or a shadow.
func validToken(k, v string) bool {
	switch {
	case !slices.Contains(ThemeTokens, k):
		return false
	case strings.HasPrefix(k, "--radius"):
		return radiusVal.MatchString(v)
	case strings.HasPrefix(k, "--font-"):
		_, ok := FontStacks[v]
		return ok
	case strings.HasPrefix(k, "--shadow"):
		return shadowVal.MatchString(v)
	case k == "--glow":
		return glowVal.MatchString(v)
	case k == "--page-gradient":
		return gradientVal.MatchString(v)
	default:
		return colorValue.MatchString(v)
	}
}

// ThemeCSSMarker starts the custom CSS written by applying a theme.
const ThemeCSSMarker = "/* trisse-home theme: "

// ThemeDef is a gallery theme.
type ThemeDef struct {
	ID     string
	Name   string
	Author string
	// Scheme is light or dark: the base the tokens are meant for. Empty for
	// a theme with variants, which follows the account's light/dark choice.
	Scheme Theme
	// Tokens apply to every scheme. In a theme with variants they hold what
	// both variants share (type, radii, shadows).
	Tokens map[string]string
	// Variants hold the tokens for light and for dark (#505), or nothing.
	Variants map[Theme]map[string]string
	// BuiltIn themes ship with the app and cannot be deleted.
	BuiltIn bool
}

// HasVariants reports whether the theme brings a light and a dark variant.
func (t ThemeDef) HasVariants() bool { return len(t.Variants) > 0 }

// ErrThemeNotFound is returned for unknown themes.
var ErrThemeNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The theme does not exist.")

// NewThemeDef validates a theme. A theme is either made for one scheme
// (scheme light or dark, all tokens in tokens) or has a light and a dark
// variant (scheme empty); docs/appearance.md describes the file format.
func NewThemeDef(id, name, author string, scheme Theme, tokens map[string]string, variants map[Theme]map[string]string) (ThemeDef, error) {
	t := ThemeDef{ID: strings.ToLower(strings.TrimSpace(id)), Name: strings.TrimSpace(name), Author: strings.TrimSpace(author), Scheme: scheme}
	switch {
	case !themeID.MatchString(t.ID):
		return ThemeDef{}, kernel.NewValidationError("id", "The theme id must be 2–40 lowercase letters, digits or '-'.")
	case t.Name == "" || utf8.RuneCountInString(t.Name) > 40:
		return ThemeDef{}, kernel.NewValidationError("name", "The theme needs a name of at most 40 characters.")
	case utf8.RuneCountInString(t.Author) > 60:
		return ThemeDef{}, kernel.NewValidationError("author", "The author must be at most 60 characters long.")
	case len(variants) == 0 && scheme != ThemeLight && scheme != ThemeDark:
		return ThemeDef{}, kernel.NewValidationError("scheme", "The scheme must be light or dark.")
	case len(variants) > 0 && scheme != "":
		return ThemeDef{}, kernel.NewValidationError("scheme", "A theme with variants has no scheme of its own.")
	case len(variants) > 0 && (len(variants) != 2 || variants[ThemeLight] == nil || variants[ThemeDark] == nil):
		return ThemeDef{}, kernel.NewValidationError("variants", "A theme with variants needs exactly a light and a dark variant.")
	}
	var err error
	if t.Tokens, err = cleanTokens(tokens); err != nil {
		return ThemeDef{}, err
	}
	if len(variants) == 0 {
		if err := requireTokens(t.Tokens); err != nil {
			return ThemeDef{}, err
		}
		return t, nil
	}
	t.Variants = make(map[Theme]map[string]string, len(variants))
	for side, vt := range variants {
		clean, err := cleanTokens(vt)
		if err != nil {
			return ThemeDef{}, err
		}
		if err := requireTokens(merge(t.Tokens, clean)); err != nil {
			return ThemeDef{}, err
		}
		t.Variants[side] = clean
	}
	return t, nil
}

// cleanTokens validates tokens and lowers the case of their values.
func cleanTokens(tokens map[string]string) (map[string]string, error) {
	out := make(map[string]string, len(tokens))
	for k, v := range tokens {
		v = strings.ToLower(strings.TrimSpace(v))
		if !slices.Contains(ThemeTokens, k) {
			return nil, kernel.NewValidationErrorf("tokens", "Unknown token {token}.", "token", k)
		}
		if !validToken(k, v) {
			return nil, kernel.NewValidationErrorf("tokens", "Invalid value for {token}.", "token", k)
		}
		out[k] = v
	}
	return out, nil
}

func requireTokens(tokens map[string]string) error {
	for _, k := range requiredTokens {
		if _, ok := tokens[k]; !ok {
			return kernel.NewValidationErrorf("tokens", "The theme must set {token}.", "token", k)
		}
	}
	return nil
}

func merge(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// Resolved returns the tokens in effect for a scheme: the shared tokens and,
// in a theme with variants, that variant on top.
func (t ThemeDef) Resolved(scheme Theme) map[string]string {
	if !t.HasVariants() {
		return t.Tokens
	}
	if scheme != ThemeDark {
		scheme = ThemeLight
	}
	return merge(t.Tokens, t.Variants[scheme])
}

// CSS is the custom CSS that applies the theme, marked so a later apply knows
// it is not the user's own CSS. A single-scheme theme sets its tokens for
// both schemes; a theme with variants sets the light ones on the root and
// the dark ones where the dark scheme is active, so it follows the account's
// light, dark or system choice.
func (t ThemeDef) CSS() string {
	var b strings.Builder
	b.WriteString(ThemeCSSMarker + t.ID + " */\n")
	writeBlock(&b, ":root:root,\n:root:root[data-theme]", t.Resolved(ThemeLight))
	if t.HasVariants() {
		writeBlock(&b, ":root:root[data-theme='dark']", t.Variants[ThemeDark])
	}
	return b.String()
}

func writeBlock(b *strings.Builder, selector string, tokens map[string]string) {
	keys := make([]string, 0, len(tokens))
	for k := range tokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b.WriteString(selector + " {\n")
	for _, k := range keys {
		b.WriteString("  " + k + ": " + cssValue(k, tokens[k]) + ";\n")
	}
	b.WriteString("}\n")
}

// cssValue turns a stored token value into CSS: a font keyword becomes its
// font stack, every other value is CSS already.
func cssValue(k, v string) string {
	if strings.HasPrefix(k, "--font-") {
		return FontStacks[v]
	}
	return v
}

// IsThemeCSS reports whether custom CSS was written by applying a theme.
func IsThemeCSS(css string) bool { return strings.HasPrefix(css, ThemeCSSMarker) }

// AppliedTheme returns the id of the theme the custom CSS comes from, or "".
func AppliedTheme(css string) string {
	if !IsThemeCSS(css) {
		return ""
	}
	rest := strings.TrimPrefix(css, ThemeCSSMarker)
	id, _, _ := strings.Cut(rest, " ")
	return id
}

// WithStyle returns the appearance with other custom CSS and accent color
// and, unless scheme is empty, another scheme. A theme brings its own accent,
// so applying one clears the accent color (it would override the theme).
func (a Appearance) WithStyle(css string, scheme Theme, accentColor string) (Appearance, error) {
	if scheme == "" {
		scheme = a.theme
	}
	next, err := NewAppearance(scheme, accentColor, a.backgroundImage, css)
	if err != nil {
		return Appearance{}, err
	}
	next.language, next.greetingName, next.launcher = a.language, a.greetingName, a.launcher
	return next, nil
}

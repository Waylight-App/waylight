package domain

import (
	"errors"
	"os"
	"strings"
	"testing"

	"waylight/internal/shared/kernel"
)

func base() map[string]string {
	return map[string]string{"--color-bg": "#2E3440", "--color-surface": "#3b4252", "--color-text": "#eceff4", "--color-accent": "#88c0d0", "--radius": "6px"}
}

func TestThemeDef(t *testing.T) {
	th, err := NewThemeDef("nord", "Nord", "Arctic Ice Studio", ThemeDark, base(), nil)
	if err != nil || th.Tokens["--color-bg"] != "#2e3440" {
		t.Fatalf("%+v %v", th, err)
	}
	css := th.CSS()
	if !strings.HasPrefix(css, "/* waylight theme: nord */") || !strings.Contains(css, "--color-accent: #88c0d0;") || !IsThemeCSS(css) || AppliedTheme(css) != "nord" {
		t.Fatalf("%s", css)
	}
	for name, mod := range map[string]func(map[string]string){
		"unknown token":   func(m map[string]string) { m["--color-unknown"] = "#fff" },
		"font as color":   func(m map[string]string) { m["--font-sans"] = "#fff" },
		"url value":       func(m map[string]string) { m["--color-bg"] = "url(https://evil.example/x.png)" },
		"css injection":   func(m map[string]string) { m["--color-bg"] = "#fff; } body { display:none" },
		"expression":      func(m map[string]string) { m["--color-accent"] = "red" },
		"radius too big":  func(m map[string]string) { m["--radius"] = "999px" },
		"radius as color": func(m map[string]string) { m["--radius"] = "#fff" },
		"missing text":    func(m map[string]string) { delete(m, "--color-text") },
	} {
		tok := base()
		mod(tok)
		if _, err := NewThemeDef("x1", "X", "", ThemeLight, tok, nil); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, id := range []string{"A", "x", "../x", "a b"} {
		if _, err := NewThemeDef(id, "X", "", ThemeLight, base(), nil); err == nil {
			t.Errorf("id %q accepted", id)
		}
	}
	if _, err := NewThemeDef("x1", "X", "", "system", base(), nil); err == nil {
		t.Error("scheme system")
	}
	if IsThemeCSS("body{}") || AppliedTheme("body{}") != "" {
		t.Error("own CSS is not a theme")
	}
}

func TestThemeTokenValues(t *testing.T) {
	good := map[string]string{
		"--font-sans":     "figtree",
		"--font-display":  "fraunces",
		"--shadow":        "0 1px 2px #1b22381a",
		"--shadow-pop":    "0 12px 40px #0000002e, 0 1px 2px -1px #00000014",
		"--glow":          "0 0 80px #f2b4412e",
		"--page-gradient": "linear-gradient(180deg, #fff3d6, #fbe8bc)",
	}
	for k, v := range good {
		if !validToken(k, v) {
			t.Errorf("%s: %q refused", k, v)
		}
	}
	bad := map[string]string{
		"--font-sans":     "url(https://evil.example/font.woff2)",
		"--font-display":  "'comic sans ms'",
		"--shadow":        "none",
		"--shadow-pop":    "0 0 1px #000000, 0 0 1px #000000, 0 0 1px #000000, 0 0 1px #000000",
		"--glow":          "0 0 80px #f2b441; } body { display:none",
		"--page-gradient": "linear-gradient(180deg, #ffffff, url(x))",
	}
	for k, v := range bad {
		if validToken(k, v) {
			t.Errorf("%s: %q accepted", k, v)
		}
	}
}

func TestThemeVariants(t *testing.T) {
	light := map[string]string{"--color-bg": "#fff3d6", "--color-surface": "#fffbf0", "--color-text": "#1b2238", "--color-accent": "#f2b441"}
	dark := map[string]string{"--color-bg": "#131829", "--color-surface": "#1b2238", "--color-text": "#f2efe8", "--color-accent": "#f2b441"}
	shared := map[string]string{"--font-display": "Fraunces", "--radius": "14px"}
	th, err := NewThemeDef("amber", "Amber", "", "", shared, map[Theme]map[string]string{ThemeLight: light, ThemeDark: dark})
	if err != nil || !th.HasVariants() {
		t.Fatalf("%+v %v", th, err)
	}
	if r := th.Resolved(ThemeDark); r["--color-bg"] != "#131829" || r["--radius"] != "14px" {
		t.Errorf("resolved dark: %v", r)
	}
	if r := th.Resolved(ThemeSystem); r["--color-bg"] != "#fff3d6" {
		t.Errorf("resolved fallback is light: %v", r)
	}
	css := th.CSS()
	for _, want := range []string{
		":root:root,\n:root:root[data-theme] {\n",
		"--color-bg: #fff3d6;",
		"--font-display: 'Fraunces Variable', Fraunces, Georgia, serif;",
		":root:root[data-theme='dark'] {\n",
		"--color-bg: #131829;",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("CSS lacks %q:\n%s", want, css)
		}
	}
	if AppliedTheme(css) != "amber" {
		t.Error("marker")
	}

	both := map[Theme]map[string]string{ThemeLight: light, ThemeDark: dark}
	noText := map[string]string{"--color-bg": "#000", "--color-surface": "#111", "--color-accent": "#fff"}
	badValue := map[string]string{"--color-bg": "url(x)", "--color-surface": "#111", "--color-text": "#fff", "--color-accent": "#fff"}
	for name, c := range map[string]struct {
		scheme   Theme
		variants map[Theme]map[string]string
	}{
		"scheme with variants": {ThemeDark, both},
		"only light":           {"", map[Theme]map[string]string{ThemeLight: light}},
		"system variant":       {"", map[Theme]map[string]string{ThemeLight: light, ThemeSystem: dark}},
		"dark misses text":     {"", map[Theme]map[string]string{ThemeLight: light, ThemeDark: noText}},
		"bad value in variant": {"", map[Theme]map[string]string{ThemeLight: light, ThemeDark: badValue}},
	} {
		if _, err := NewThemeDef("x1", "X", "", c.scheme, shared, c.variants); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// The gallery preview in the web app keeps its own copy of the font stacks;
// both must stay the same, or a card would preview another font than the
// applied theme uses.
func TestFontStacksMatchWeb(t *testing.T) {
	src, err := os.ReadFile("../../../../web/src/features/appearance/domain/themes.ts")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range FontStacks {
		if want := k + `: "` + v + `",`; !strings.Contains(string(src), want) {
			t.Errorf("themes.ts lacks %s", want)
		}
	}
}

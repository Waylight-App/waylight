package gateway

import (
	"math"
	"strconv"
	"testing"

	"waylight/internal/modules/appearance/domain"
)

// luminance is the WCAG relative luminance of #rgb or #rrggbb.
func luminance(hex string) float64 {
	h := hex[1:]
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	var l float64
	for i, w := range []float64{0.2126, 0.7152, 0.0722} {
		v, _ := strconv.ParseUint(h[2*i:2*i+2], 16, 8)
		c := float64(v) / 255
		if c <= 0.04045 {
			c /= 12.92
		} else {
			c = math.Pow((c+0.055)/1.055, 2.4)
		}
		l += w * c
	}
	return l
}

func contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	return (math.Max(la, lb) + 0.05) / (math.Min(la, lb) + 0.05)
}

// Every built-in theme meets WCAG AA (4.5:1) for all text it is used for (#113).
// The accent itself is a fill with --color-accent-contrast on it; where it
// appears as text, icon or focus ring, --color-accent-text is used (#504,
// #506), which falls back to the accent and is checked below.
func TestBuiltinThemesMeetWCAGAA(t *testing.T) {
	pairs := map[string][]string{
		"--color-text":            {"--color-bg", "--color-surface", "--color-surface-2"},
		"--color-text-muted":      {"--color-bg", "--color-surface", "--color-surface-2"},
		"--color-accent-text":     {"--color-surface", "--color-surface-2"},
		"--color-danger":          {"--color-surface"},
		"--color-accent-contrast": {"--color-accent"},
	}
	for _, th := range BuiltinThemes() {
		schemes := []domain.Theme{th.Scheme}
		if th.HasVariants() {
			schemes = []domain.Theme{domain.ThemeLight, domain.ThemeDark}
		}
		for _, scheme := range schemes {
			tokens := th.Resolved(scheme)
			// A theme without its own accent text colour uses the accent (#504).
			token := func(name string) (string, bool) {
				if v, ok := tokens[name]; ok {
					return v, true
				}
				if name == "--color-accent-text" {
					v, ok := tokens["--color-accent"]
					return v, ok
				}
				return "", false
			}
			for fg, bgs := range pairs {
				for _, bg := range bgs {
					a, okA := token(fg)
					b, okB := token(bg)
					if !okA || !okB {
						continue
					}
					if c := contrast(a, b); c < 4.5 {
						t.Errorf("%s (%s): %s %s on %s %s = %.2f:1, want 4.5:1", th.ID, scheme, fg, a, bg, b, c)
					}
				}
			}
		}
	}
}

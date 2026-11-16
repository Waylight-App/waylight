// Package domain holds the rules of an organisation's branding (#305).
//
// Ring: Entities.
package domain

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Settings are the organisation's name, colours and whether the product
// credit is shown.
type Settings struct {
	Name              string
	Accent, Secondary string // "#rrggbb" or empty
	HidePoweredBy     bool
}

// Limits.
const (
	MaxName = 40
	// MaxUpload is the largest image accepted before re-encoding.
	MaxUpload = 2 << 20
	// MaxEdge is the longest edge an uploaded image is scaled to.
	MaxEdge = 512
)

var hexColour = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Validate checks and normalises the settings.
func (s *Settings) Validate() error {
	s.Name = strings.TrimSpace(s.Name)
	s.Accent, s.Secondary = strings.ToLower(strings.TrimSpace(s.Accent)), strings.ToLower(strings.TrimSpace(s.Secondary))
	switch {
	case utf8.RuneCountInString(s.Name) > MaxName:
		return kernel.NewValidationError("name", "The name must be at most 40 characters.")
	case s.Accent != "" && !hexColour.MatchString(s.Accent):
		return kernel.NewValidationError("accent", "Enter the accent colour as #rrggbb.")
	case s.Secondary != "" && !hexColour.MatchString(s.Secondary):
		return kernel.NewValidationError("secondary", "Enter the secondary colour as #rrggbb.")
	}
	return nil
}

// Empty reports whether nothing is set, so the product's branding applies.
func (s Settings) Empty() bool {
	return s.Name == "" && s.Accent == "" && s.Secondary == "" && !s.HidePoweredBy
}

// Backgrounds the accent colour is checked against: the light and the dark
// default theme.
var Backgrounds = map[string]string{"light": "#ffffff", "dark": "#0f172a"}

// MinContrast is WCAG AA for normal text.
const MinContrast = 4.5

func luminance(hex string) float64 {
	channel := func(i int) float64 {
		v, _ := strconv.ParseUint(hex[i:i+2], 16, 8)
		c := float64(v) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(1) + 0.7152*channel(3) + 0.0722*channel(5)
}

// Contrast is the WCAG contrast ratio of two "#rrggbb" colours.
func Contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// Warnings lists where a colour falls below WCAG AA, for the settings page.
func (s Settings) Warnings() []string {
	var out []string
	for _, c := range []struct{ field, colour string }{{"accent", s.Accent}, {"secondary", s.Secondary}} {
		if c.colour == "" {
			continue
		}
		for _, theme := range []string{"light", "dark"} {
			if r := Contrast(c.colour, Backgrounds[theme]); r < MinContrast {
				out = append(out, fmt.Sprintf("%s on the %s theme: contrast %.1f:1, below 4.5:1", c.field, theme, r))
			}
		}
	}
	return out
}

// Errors.
var (
	ErrNotAllowed = kernel.NewPublicError(kernel.ErrForbidden, "Only an admin can change the branding.")
	ErrKind       = kernel.NewPublicError(kernel.ErrNotFound, "Upload a logo, a dark logo or a favicon.")
	ErrImage      = kernel.NewValidationError("image", "Upload a PNG, JPEG or GIF of at most 2 MB; SVG is not accepted.")
)

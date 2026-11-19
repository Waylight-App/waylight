package answers

import (
	"fmt"
	"strings"
)

// Lang is a language of the answers.
type Lang string

// The languages; English is the fallback.
const (
	English Lang = "en"
	German  Lang = "de"
)

// ParseLang picks a supported language from a tag such as "de-DE" or an
// Accept-Language header; English otherwise.
func ParseLang(tag string) Lang {
	for part := range strings.SplitSeq(tag, ",") {
		code, _, _ := strings.Cut(strings.TrimSpace(part), ";")
		code, _, _ = strings.Cut(strings.ToLower(code), "-")
		switch Lang(code) {
		case German:
			return German
		case English:
			return English
		}
	}
	return English
}

// Messages are the texts of a source by language and key; the English text
// is the fallback for a missing translation.
type Messages map[Lang]map[string]string

// T returns the text of key in lang with "{name}" placeholders filled from
// args, given as name, value pairs.
func (m Messages) T(lang Lang, key string, args ...any) string {
	s, ok := m[lang][key]
	if !ok {
		s, ok = m[English][key]
	}
	if !ok {
		return key
	}
	for i := 0; i+1 < len(args); i += 2 {
		s = strings.ReplaceAll(s, "{"+fmt.Sprint(args[i])+"}", fmt.Sprint(args[i+1]))
	}
	return s
}

// common are the texts of the search itself (asking back, failures).
var common = Messages{
	English: {
		"which_title":       "Which one?",
		"which":             "Which one do you mean?",
		"unavailable_title": "Not available right now",
		"unavailable":       "This information cannot be loaded right now. Try again later or open the widget.",
	},
	German: {
		"which_title":       "Welcher?",
		"which":             "Welchen meinst du?",
		"unavailable_title": "Gerade nicht verfügbar",
		"unavailable":       "Diese Information lässt sich gerade nicht laden. Versuch es später noch einmal oder öffne das Widget.",
	},
}

// FormatBytes formats a size in binary units, e.g. "1.5 GiB".
func FormatBytes(bytes uint64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	value := float64(bytes)
	unit := 0
	for value >= 1024 && unit < len(units)-1 {
		value /= 1024
		unit++
	}
	if value >= 10 || unit == 0 {
		return fmt.Sprintf("%.0f %s", value, units[unit])
	}
	return fmt.Sprintf("%.1f %s", value, units[unit])
}

// FormatPercent rounds a share to a whole percent, e.g. "42 %".
func FormatPercent(p float64) string { return fmt.Sprintf("%.0f %%", p) }

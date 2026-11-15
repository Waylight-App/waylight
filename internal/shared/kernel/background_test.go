package kernel

import (
	"errors"
	"testing"
)

func TestParseBackground(t *testing.T) {
	for _, ok := range []string{"", "  ", "https://img.example/a.jpg", "builtin:dusk", "/api/appearance/backgrounds/abc-123"} {
		if _, err := ParseBackground("background", ok); err != nil {
			t.Errorf("%q rejected: %v", ok, err)
		}
	}
	for _, bad := range []string{
		"ftp://x/a.jpg", "javascript:alert(1)", "https://x/a.jpg) ; color: red", "/etc/passwd",
		"/api/appearance/backgrounds/../x", "builtin:Dusk!", "data:image/png;base64,AAAA",
	} {
		if _, err := ParseBackground("background", bad); !errors.Is(err, ErrInvalid) {
			t.Errorf("%q accepted", bad)
		}
	}
}

package domain

import (
	"math"
	"testing"
)

func TestContrastAndWarnings(t *testing.T) {
	if r := Contrast("#000000", "#ffffff"); math.Abs(r-21) > 0.01 {
		t.Fatalf("black on white %.2f", r)
	}
	s := Settings{Accent: "#FFEE00"}
	if err := s.Validate(); err != nil || s.Accent != "#ffee00" {
		t.Fatal(err, s.Accent)
	}
	if w := s.Warnings(); len(w) != 1 {
		t.Fatalf("yellow should fail on light only: %v", w)
	}
	if w := (Settings{Accent: "#1d4ed8"}).Warnings(); len(w) != 1 {
		t.Fatalf("dark blue should fail on dark only: %v", w)
	}
	for _, bad := range []Settings{{Accent: "red"}, {Secondary: "#12345"}, {Name: "12345678901234567890123456789012345678901"}} {
		if bad.Validate() == nil {
			t.Errorf("accepted %+v", bad)
		}
	}
}

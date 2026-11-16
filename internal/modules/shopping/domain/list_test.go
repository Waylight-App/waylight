package domain

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestExportICSEscapesAndFolds(t *testing.T) {
	long := strings.Repeat("Äpfel und Birnen, ", 10)
	items := []Item{{ID: "a", Summary: long}, {ID: "b", Summary: "Milk; 1 l", Done: true, Completed: time.Unix(0, 0)}}
	out := ExportICS(List{Name: "Groceries, family"}, items, "home.example", time.Unix(0, 0))
	for _, line := range strings.Split(strings.TrimSuffix(out, "\r\n"), "\r\n") {
		if len(line) > 75 {
			t.Fatalf("line longer than 75 octets: %q", line)
		}
		if !utf8.ValidString(line) {
			t.Fatalf("folded inside a character: %q", line)
		}
	}
	unfolded := strings.ReplaceAll(out, "\r\n ", "")
	if !strings.Contains(unfolded, `X-WR-CALNAME:Groceries\, family`) || !strings.Contains(unfolded, `SUMMARY:Milk\; 1 l`) ||
		!strings.Contains(unfolded, "SUMMARY:"+strings.ReplaceAll(long, ",", `\,`)) || !strings.Contains(unfolded, "UID:a@home.example") {
		t.Fatalf("content: %s", unfolded)
	}
}

func TestListName(t *testing.T) {
	if n, err := NormaliseListName("  Hardware   store "); err != nil || n != "Hardware store" {
		t.Fatalf("%q %v", n, err)
	}
	if _, err := NormaliseListName("   "); err == nil {
		t.Fatal("empty name accepted")
	}
}

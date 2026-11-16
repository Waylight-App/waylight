package domain

import (
	"strings"
	"testing"
)

func TestResultsEscapeAndLimit(t *testing.T) {
	hits := []Hit{{Source: "Service", Title: "VPN <admin> & co | x", URL: "https://vpn.example/a>b"}}
	for range 10 {
		hits = append(hits, Hit{Source: "Bookmark", Title: "b", URL: "https://b.example"})
	}
	got := Results("vpn <x>", hits, false)
	if !strings.Contains(got, "“vpn &lt;x&gt;”") || !strings.Contains(got, "<https://vpn.example/a%3Eb|VPN &lt;admin&gt; &amp; co ¦ x> · Service") {
		t.Fatalf("escaping: %s", got)
	}
	if strings.Count(got, "\n• ") != MaxHits || !strings.HasSuffix(got, "More on the start page.") {
		t.Fatalf("limit: %s", got)
	}
	if Results("x", nil, false) != "Nothing found for “x”." {
		t.Fatal("empty")
	}
}

package domain

import (
	"strings"
	"testing"
)

func TestNewKeyword(t *testing.T) {
	ok := []struct {
		word   string
		kind   TargetKind
		target string
	}{
		{" NAS ", TargetService, "svc-1"},
		{"bank", TargetURL, "https://bank.example/login"},
		{"wetter", TargetWidget, "w_42"},
		{"müll", TargetQuery, "waste next week"},
		{"git-hub_2", TargetBookmark, "b9"},
	}
	for _, c := range ok {
		k, err := NewKeyword(c.word, c.kind, c.target, "")
		if err != nil {
			t.Errorf("%q: %v", c.word, err)
			continue
		}
		if k.Word() != strings.ToLower(strings.TrimSpace(c.word)) {
			t.Errorf("word %q not normalised: %q", c.word, k.Word())
		}
	}

	bad := []struct {
		word   string
		kind   TargetKind
		target string
		label  string
	}{
		{"", TargetURL, "https://x.example", ""},
		{"!yt", TargetURL, "https://x.example", ""},
		{"two words", TargetURL, "https://x.example", ""},
		{"-dash", TargetURL, "https://x.example", ""},
		{strings.Repeat("a", 25), TargetURL, "https://x.example", ""},
		{"nas", TargetURL, "ftp://x.example", ""},
		{"nas", TargetURL, "javascript:alert(1)", ""},
		{"nas", TargetURL, "https://", ""},
		{"nas", TargetService, "", ""},
		{"nas", TargetService, "has space", ""},
		{"nas", "app", "x", ""},
		{"nas", TargetQuery, "x", strings.Repeat("ä", 101)},
	}
	for _, c := range bad {
		if _, err := NewKeyword(c.word, c.kind, c.target, c.label); err == nil {
			t.Errorf("NewKeyword(%q, %q, %q) accepted invalid input", c.word, c.kind, c.target)
		}
	}
}

func TestCheckKeywordConflict(t *testing.T) {
	p := CapturePrefixes{Todo: "t", Note: "n", Bookmark: "bm"}
	if err := CheckKeywordConflict("bm", p); err == nil || !strings.Contains(err.Error(), "bookmarks") {
		t.Errorf("want a clear conflict with the bookmark prefix, got %v", err)
	}
	if err := CheckKeywordConflict("nas", p); err != nil {
		t.Errorf("unexpected conflict: %v", err)
	}
}

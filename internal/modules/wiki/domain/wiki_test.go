package domain

import (
	"errors"
	"strings"
	"testing"

	"waylight/internal/shared/kernel"
)

func TestValidate(t *testing.T) {
	w := Wiki{Name: " Handbook ", Kind: KindBookStack, URL: "https://wiki.example.org/", Token: " id:secret "}
	if err := w.Validate(); err != nil || w.Name != "Handbook" || w.URL != "https://wiki.example.org" || w.Token != "id:secret" {
		t.Fatalf("%+v %v", w, err)
	}
	for _, bad := range []Wiki{
		{Name: "", Kind: KindOutline, URL: "https://x.example", Token: "t"},
		{Name: "x", Kind: "confluence", URL: "https://x.example", Token: "t"},
		{Name: "x", Kind: KindWikiJS, URL: "ftp://x.example", Token: "t"},
		{Name: "x", Kind: KindWikiJS, URL: "https://x.example", Token: ""},
	} {
		if err := bad.Validate(); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%+v accepted: %v", bad, err)
		}
	}
}

func TestNormaliseQuery(t *testing.T) {
	if _, ok := NormaliseQuery(" a "); ok {
		t.Fatal("one letter is not sent")
	}
	if q, ok := NormaliseQuery("  vpn   setup "); !ok || q != "vpn setup" {
		t.Fatalf("%q", q)
	}
	if q, _ := NormaliseQuery(strings.Repeat("x", 500)); len(q) != MaxQueryLength {
		t.Fatalf("%d", len(q))
	}
}

func TestExcerpt(t *testing.T) {
	got := Excerpt(`<p>How to set up <strong>VPN</strong> &amp; DNS</p>`)
	if got != "How to set up VPN & DNS" {
		t.Fatalf("%q", got)
	}
	if e := Excerpt(strings.Repeat("word ", 200)); len([]rune(e)) != maxExcerptLength {
		t.Fatalf("%d", len([]rune(e)))
	}
}

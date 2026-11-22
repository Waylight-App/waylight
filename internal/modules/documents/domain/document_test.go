package domain

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

var png = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

func TestStatusOn(t *testing.T) {
	day := time.Date(2026, 11, 22, 15, 0, 0, 0, time.UTC)
	for end, want := range map[string]Status{
		"":           StatusNone,
		"2026-11-21": StatusEnded,
		"2026-11-22": EndingSoon, // the end day still counts as covered
		"2026-12-22": EndingSoon,
		"2026-12-23": Running,
	} {
		d, err := New("1", Input{Title: "TV", WarrantyEnd: end}, kernel.Access{}, day)
		if err != nil {
			t.Fatal(err)
		}
		if got := d.StatusOn(day); got != want {
			t.Errorf("end %q: %s, want %s", end, got, want)
		}
	}
}

func TestInputRules(t *testing.T) {
	for name, in := range map[string]Input{
		"no title":      {},
		"long title":    {Title: strings.Repeat("a", 81)},
		"bad day":       {Title: "TV", PurchasedOn: "yesterday"},
		"end before":    {Title: "TV", PurchasedOn: "2026-05-01", WarrantyEnd: "2026-04-30"},
		"long note":     {Title: "TV", Note: strings.Repeat("a", 501)},
		"bad warranty":  {Title: "TV", WarrantyEnd: "2026-13-01"},
		"whitespace":    {Title: "  "},
		"day with time": {Title: "TV", WarrantyEnd: "2026-05-01T10:00"},
	} {
		if _, err := New("1", in, kernel.Access{}, time.Now()); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := New("1", Input{Title: " TV ", PurchasedOn: "2026-05-01", WarrantyEnd: "2026-05-01"}, kernel.Access{}, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestNewFile(t *testing.T) {
	f, err := NewFile(`..\evil/../receipt.png`, png)
	if err != nil || f.ContentType != "image/png" || f.Name != "receipt.png" || f.Size != int64(len(png)) {
		t.Fatalf("%+v %v", f, err)
	}
	if f, err := NewFile("a.pdf", []byte("%PDF-1.4\n")); err != nil || f.ContentType != "application/pdf" {
		t.Fatalf("pdf: %+v %v", f, err)
	}
	// The type comes from the bytes: a script named like a photo is refused.
	for name, data := range map[string][]byte{
		"empty":  nil,
		"script": []byte("<script>alert(1)</script>"),
		"text":   []byte("plain text"),
		"huge":   append(png, bytes.Repeat([]byte{0}, MaxFileBytes)...),
	} {
		if _, err := NewFile("photo.png", data); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if got := CleanFileName(`a"b`); got != "ab" {
		t.Errorf("quote kept: %q", got)
	}
	if got := CleanFileName(".."); got != "file" {
		t.Errorf("dots: %q", got)
	}
}

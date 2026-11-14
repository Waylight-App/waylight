package domain

import (
	"errors"
	"strings"
	"testing"

	"trisse-home/internal/shared/kernel"
)

func TestNormaliseAlbum(t *testing.T) {
	if a, _ := NormaliseAlbum("  "); a != DefaultAlbum {
		t.Fatalf("empty → default, got %q", a)
	}
	if a, _ := NormaliseAlbum(" Summer   2026 "); a != "Summer 2026" {
		t.Fatalf("got %q", a)
	}
	if _, err := NormaliseAlbum(strings.Repeat("x", 41)); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("too long")
	}
}

func TestCheckOrder(t *testing.T) {
	ps := []Photo{{ID: "a"}, {ID: "b"}}
	if err := CheckOrder(ps, []string{"b", "a"}); err != nil {
		t.Fatal(err)
	}
	for _, ids := range [][]string{{"a"}, {"a", "a"}, {"a", "c"}} {
		if CheckOrder(ps, ids) == nil {
			t.Fatalf("%v accepted", ids)
		}
	}
}

func TestUsageFits(t *testing.T) {
	if !(Usage{}).Fits(1 << 20) {
		t.Fatal("empty storage")
	}
	if (Usage{Bytes: QuotaBytes - 10}).Fits(11) || (Usage{Count: MaxPhotos}).Fits(1) {
		t.Fatal("over the limit")
	}
}

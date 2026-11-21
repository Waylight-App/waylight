package domain

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

var (
	now   = time.Date(2026, 11, 23, 18, 0, 0, 0, time.UTC)
	today = now
)

func TestNew(t *testing.T) {
	n, err := New("n1", Input{Text: " Dentist on Friday ", ExpiresOn: "2026-11-27"}, kernel.Access{}, now, today)
	if err != nil || n.Text != "Dentist on Friday" || n.ExpiryDate() != "2026-11-27" || n.ExpiredOn(today) {
		t.Fatalf("%+v %v", n, err)
	}
	if !n.ExpiredOn(time.Date(2026, 11, 28, 0, 30, 0, 0, time.UTC)) {
		t.Fatal("should be expired the day after")
	}
	for name, in := range map[string]Input{
		"empty":    {Text: " "},
		"long":     {Text: strings.Repeat("x", 501)},
		"bad date": {Text: "x", ExpiresOn: "Friday"},
		"past":     {Text: "x", ExpiresOn: "2026-11-22"},
	} {
		if _, err := New("n", in, kernel.Access{}, now, today); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}
	forever, _ := New("n2", Input{Text: "Wifi password is on the fridge"}, kernel.Access{}, now, today)
	if forever.ExpiryDate() != "" || forever.ExpiredOn(now.AddDate(10, 0, 0)) {
		t.Fatal("a note without expiry never expires")
	}
}

func TestLess(t *testing.T) {
	list := []Note{
		{ID: "old", CreatedAt: now.Add(-time.Hour)},
		{ID: "pinned-old", Pinned: true, CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "new", CreatedAt: now},
	}
	slices.SortFunc(list, func(a, b Note) int {
		if Less(a, b) {
			return -1
		}
		if Less(b, a) {
			return 1
		}
		return 0
	})
	if list[0].ID != "pinned-old" || list[1].ID != "new" || list[2].ID != "old" {
		t.Fatalf("%v", list)
	}
}

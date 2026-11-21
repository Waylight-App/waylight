package domain

import (
	"testing"
	"time"
)

func TestSortPutsFailuresFirstThenOldest(t *testing.T) {
	t0 := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	items := []Item{
		{ID: "a", Kind: KindSetup, Since: t0},
		{ID: "b", Kind: KindConnection, Since: t0.Add(time.Minute)},
		{ID: "c", Kind: KindConnection, Since: t0},
		{ID: "d", Kind: KindSetup, Since: t0.Add(-time.Hour)},
	}
	Sort(items)
	got := ""
	for _, it := range items {
		got += it.ID
	}
	if got != "cbda" {
		t.Fatalf("order = %s, want cbda", got)
	}
}

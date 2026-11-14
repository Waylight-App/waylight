package domain

import (
	"math"
	"testing"
	"time"
)

func TestSymbols(t *testing.T) {
	for in, want := range map[string]string{"fx:usd": "fx:USD", "stooq:AAPL.US": "stooq:aapl.us", "\tstooq:^dax\n": "stooq:^dax", "FX:CHF": "fx:CHF"} {
		s, err := ParseSymbol(in)
		if err != nil || s.String() != want {
			t.Errorf("%q: %v %v", in, s, err)
		}
	}
	for _, bad := range []string{"usd", "fx:EUR", "fx:DOLLAR", "stooq:", "yahoo:aapl", "stooq:a/b"} {
		if _, err := ParseSymbol(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
	list, err := ParseSymbols([]string{"fx:USD", "fx:usd", "stooq:aapl.us"})
	if err != nil || len(list) != 2 {
		t.Fatal(list, err)
	}
	many := make([]string, 13)
	for i := range many {
		many[i] = "stooq:s" + string(rune('a'+i))
	}
	if _, err := ParseSymbols(many); err == nil {
		t.Fatal("13 entries accepted")
	}
}

func TestNewQuote(t *testing.T) {
	d := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	q, ok := NewQuote(Symbol{SourceFX, "USD"}, []Point{{d, 1.08}, {d.AddDate(0, 0, 1), 1.0854}})
	if !ok || q.Value != 1.0854 || math.Abs(q.ChangePct-0.5) > 1e-9 || !q.AsOf.Equal(d.AddDate(0, 0, 1)) {
		t.Fatalf("%+v", q)
	}
	if _, ok := NewQuote(Symbol{}, nil); ok {
		t.Fatal("empty series")
	}
}

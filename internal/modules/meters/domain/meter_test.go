package domain

import (
	"math"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func day(s string) time.Time {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return d
}

func TestMeterInput(t *testing.T) {
	ok := Input{Name: " Electricity ", Kind: "electricity", Unit: "kWh"}
	m, err := New("1", ok, kernel.Access{}, time.Time{})
	if err != nil || m.Name != "Electricity" || m.Kind != Electricity {
		t.Fatalf("%+v %v", m, err)
	}
	for _, in := range []Input{
		{Name: "", Kind: "gas", Unit: "m³"},
		{Name: "Gas", Kind: "oil", Unit: "m³"},
		{Name: "Gas", Kind: "gas", Unit: ""},
		{Name: "Gas", Kind: "gas", Unit: "0123456789abc"},
	} {
		if _, err := New("1", in, kernel.Access{}, time.Time{}); err == nil {
			t.Errorf("accepted %+v", in)
		}
	}
}

func TestReadingChecks(t *testing.T) {
	today := day("2026-11-22")
	for _, in := range []ReadingInput{
		{Day: "22.11.2026", Value: 1}, {Day: "2026-11-23", Value: 1}, {Day: "1999-01-01", Value: 1},
		{Day: "2026-11-01", Value: -1}, {Day: "2026-11-01", Value: math.NaN()}, {Day: "2026-11-01", Value: MaxValue + 1},
	} {
		if _, err := NewReading(in, today); err == nil {
			t.Errorf("accepted %+v", in)
		}
	}
	if r, err := NewReading(ReadingInput{Day: "2026-11-22", Value: 0}, today); err != nil || r.Value != 0 {
		t.Fatalf("today: %+v %v", r, err)
	}

	existing := []Reading{{day("2026-01-01"), 100}, {day("2026-03-01"), 200}}
	for _, c := range []struct {
		r  Reading
		ok bool
	}{
		{Reading{day("2026-02-01"), 150}, true},
		{Reading{day("2026-02-01"), 99}, false},  // lower than an earlier one
		{Reading{day("2026-02-01"), 201}, false}, // higher than a later one
		{Reading{day("2026-03-01"), 120}, true},  // replaces the reading of its day
		{Reading{day("2026-03-01"), 99}, false},
		{Reading{day("2026-04-01"), 200}, true},
	} {
		if err := CheckOrder(existing, c.r); (err == nil) != c.ok {
			t.Errorf("%+v: %v", c.r, err)
		}
	}
}

func TestConsumptionInterpolatesAtTheTurnOfTheYear(t *testing.T) {
	readings := []Reading{{day("2025-12-16"), 1000}, {day("2026-01-16"), 1031}, {day("2026-02-01"), 1047}, {day("2027-01-01"), 1447}}
	c := ConsumptionOf(readings, 2026)
	// 1 Jan lies 16 of 31 days between the first two readings: 1016.
	if !c.Total.Known || !c.Complete || math.Abs(c.Total.Value-(1447-1016)) > 1e-9 {
		t.Fatalf("total %+v complete %v", c.Total, c.Complete)
	}
	if jan := c.Months[0]; !jan.Known || math.Abs(jan.Value-(1047-1016)) > 1e-9 {
		t.Fatalf("january %+v", jan)
	}
	if feb := c.Months[1]; !feb.Known || feb.Value < 0 {
		t.Fatalf("february %+v", feb)
	}
	var sum float64
	for _, m := range c.Months {
		sum += m.Value
	}
	if math.Abs(sum-c.Total.Value) > 1e-9 {
		t.Fatalf("months add up to %v, total %v", sum, c.Total.Value)
	}
}

func TestConsumptionOfAPartialYear(t *testing.T) {
	readings := []Reading{{day("2026-03-10"), 100}, {day("2026-06-10"), 190}}
	c := ConsumptionOf(readings, 2026)
	if c.Complete || !c.Total.Known || c.Total.Value != 90 {
		t.Fatalf("%+v", c)
	}
	// Only April and May lie fully between the readings.
	for i, m := range c.Months {
		if want := i == 3 || i == 4; m.Known != want {
			t.Errorf("month %d known=%v", i+1, m.Known)
		}
	}
	if c := ConsumptionOf(readings, 2025); c.Total.Known {
		t.Fatalf("a year before the readings: %+v", c.Total)
	}
	if c := ConsumptionOf(nil, 2026); c.Total.Known || c.Months[0].Known {
		t.Fatalf("no readings: %+v", c)
	}
	if got := Years(readings); len(got) != 1 || got[0] != 2026 {
		t.Fatalf("years %v", got)
	}
	if got := Years([]Reading{{day("2024-05-01"), 1}, {day("2026-05-01"), 2}}); len(got) != 3 || got[0] != 2026 || got[2] != 2024 {
		t.Fatalf("years %v", got)
	}
}

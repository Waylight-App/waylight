package domain

import (
	"testing"
	"time"
)

func TestParseZone(t *testing.T) {
	if z, err := ParseZone(""); err != nil || z != ZoneDELU {
		t.Fatal(z, err)
	}
	if _, err := ParseZone("FR"); err == nil {
		t.Fatal("unsupported zone accepted")
	}
}

func TestHourly(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	day := time.Date(2026, 11, 14, 0, 0, 0, 0, loc)
	var slots []Slot
	// Four quarter hours per hour; hour h costs 100+h €/MWh, except 03:00 which is negative.
	for h := range 24 {
		for q := range 4 {
			start := day.Add(time.Duration(h)*time.Hour + time.Duration(q)*15*time.Minute)
			p := 100 + float64(h) + float64(q) // average 101.5 + h
			if h == 3 {
				p = -20 + float64(q)
			}
			slots = append(slots, Slot{Start: start.UTC(), End: start.Add(15 * time.Minute).UTC(), EurPerMWh: p})
		}
	}
	hours := Hourly(slots, loc, 3)
	if len(hours) != 24 || !hours[0].Start.Equal(day) {
		t.Fatalf("%d hours, first %v", len(hours), hours[0].Start)
	}
	if hours[1].CentPerKWh != 10.25 || hours[3].CentPerKWh != -1.85 {
		t.Fatalf("averages: %v %v", hours[1].CentPerKWh, hours[3].CentPerKWh)
	}
	var cheapest []int
	for i, h := range hours {
		if h.Cheapest {
			cheapest = append(cheapest, i)
		}
	}
	if len(cheapest) != 3 || cheapest[0] != 0 || cheapest[1] != 1 || cheapest[2] != 3 {
		t.Fatalf("cheapest hours: %v", cheapest)
	}
	// Two days: the cheapest hours are marked per day.
	tomorrow := []Slot{{Start: day.Add(24 * time.Hour), EurPerMWh: 500}, {Start: day.Add(25 * time.Hour), EurPerMWh: 400}}
	two := Hourly(append(slots, tomorrow...), loc, 1)
	if !two[25].Cheapest || two[24].Cheapest {
		t.Fatalf("per day: %+v %+v", two[24], two[25])
	}
}

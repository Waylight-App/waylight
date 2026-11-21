package usecase

import (
	"testing"
	"time"

	"waylight/internal/modules/electricity/domain"
	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	loc := time.UTC
	day := time.Date(2026, 11, 21, 0, 0, 0, 0, loc)
	at := func(h int) time.Time { return day.Add(time.Duration(h) * time.Hour) }
	hours := []domain.Hour{
		{Start: at(10), CentPerKWh: 27.6},
		{Start: at(11), CentPerKWh: 30},
		{Start: at(14), CentPerKWh: 12, Cheapest: true},
		{Start: at(26), CentPerKWh: 8, Cheapest: true},
	}
	now := at(10).Add(20 * time.Minute)
	s := CompactState(Prices{Hours: hours}, now, loc)
	if s.Title.String() != "28 ct now" || s.Detail.String() != "Cheap from 14:00" || s.Level != kernel.CompactNone {
		t.Fatalf("before the cheap hours: %+v", s)
	}
	if s := CompactState(Prices{Hours: hours}, at(14), loc); s.Detail.String() != "Cheap now" {
		t.Fatalf("in a cheap hour: %+v", s)
	}
	if s := CompactState(Prices{Hours: hours[1:2]}, at(11), loc); s.Detail.Template != "" {
		t.Fatalf("no cheap hour left: %+v", s)
	}
	tomorrow := []domain.Hour{hours[1], hours[3]}
	if s := CompactState(Prices{Hours: tomorrow}, at(11), loc); s.Detail.String() != "Cheap tomorrow from 02:00" {
		t.Fatalf("cheap tomorrow: %+v", s)
	}
	if s := CompactState(Prices{Hours: hours, Stale: true}, now, loc); s.Level != kernel.CompactAttention || s.Detail.String() != "Prices could not be updated" {
		t.Fatalf("stale: %+v", s)
	}
	if s := CompactState(Prices{Hours: hours}, at(20), loc); s.Title.String() != "No price for this hour" {
		t.Fatalf("gap: %+v", s)
	}
}

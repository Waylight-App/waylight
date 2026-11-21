package usecase

import (
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	ev := func(summary string, start time.Time, d time.Duration, allDay bool) EventView {
		return EventView{Summary: summary, Start: start, End: start.Add(d), AllDay: allDay}
	}
	if s := CompactState(AgendaResponse{}, now, time.UTC); s.Title.String() != "No upcoming events" || s.Level != kernel.CompactNone {
		t.Fatalf("empty: %+v", s)
	}
	cases := []struct {
		events []EventView
		title  string
		detail string
	}{
		{[]EventView{ev("Dentist", now.Add(4*time.Hour), time.Hour, false), ev("Done", now.Add(-2*time.Hour), time.Hour, false)}, "Dentist", "Today at 14:00"},
		{[]EventView{ev("Standup", now.Add(-10*time.Minute), time.Hour, false)}, "Standup", "Now"},
		{[]EventView{ev("Trip", time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC), 24*time.Hour, true)}, "Trip", "Tomorrow"},
		{[]EventView{ev("Call", time.Date(2026, 11, 22, 9, 30, 0, 0, time.UTC), time.Hour, false)}, "Call", "Tomorrow at 09:30"},
		{[]EventView{ev("Party", time.Date(2026, 11, 28, 19, 0, 0, 0, time.UTC), time.Hour, false)}, "Party", "On 2026-11-28 at 19:00"},
	}
	for _, c := range cases {
		s := CompactState(AgendaResponse{Events: c.events}, now, time.UTC)
		if s.Title.String() != c.title || s.Detail.String() != c.detail {
			t.Errorf("%s: got %q / %q, want %q", c.title, s.Title.String(), s.Detail.String(), c.detail)
		}
	}
}

package domain

import (
	"testing"
	"time"
)

func TestNewReminder(t *testing.T) {
	if r := DefaultReminder(); r.Enabled || r.At() != "19:00" {
		t.Fatalf("default must be off at 19:00: %+v", r)
	}
	r, err := NewReminder(true, "18:30")
	if err != nil || !r.Enabled || r.Minute != 18*60+30 || r.At() != "18:30" {
		t.Fatalf("%+v %v", r, err)
	}
	for _, at := range []string{"", "7:00", "24:00", "19:60", "evening"} {
		if _, err := NewReminder(true, at); err == nil {
			t.Fatalf("%q must be rejected", at)
		}
	}
}

func TestReminderDue(t *testing.T) {
	loc := time.FixedZone("CET", 3600)
	r := Reminder{Enabled: true, Minute: 19 * 60}
	tomorrow := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	if _, due := r.Due(time.Date(2026, 11, 11, 18, 59, 0, 0, loc), time.Time{}); due {
		t.Fatal("not before the time")
	}
	day, due := r.Due(time.Date(2026, 11, 11, 19, 0, 0, 0, loc), time.Time{})
	if !due || !day.Equal(tomorrow) {
		t.Fatalf("due at the time for tomorrow: %v %v", day, due)
	}
	if _, due := r.Due(time.Date(2026, 11, 11, 22, 0, 0, 0, loc), tomorrow); due {
		t.Fatal("once per collection day")
	}
	if _, due := r.Due(time.Date(2026, 11, 11, 22, 0, 0, 0, loc), tomorrow.AddDate(0, 0, -1)); !due {
		t.Fatal("a server that was down sends later that evening")
	}
	if _, due := (Reminder{Minute: 19 * 60}).Due(time.Date(2026, 11, 11, 20, 0, 0, 0, loc), time.Time{}); due {
		t.Fatal("off by default")
	}
}

func TestReminderText(t *testing.T) {
	day := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	types := []Upcoming{
		{Style: Style{Label: "Yellow bag"}, Dates: []time.Time{day, day.AddDate(0, 0, 14)}},
		{Style: Style{Label: "Paper"}, Dates: []time.Time{day}},
		{Style: Style{Label: "Organic"}, Dates: []time.Time{day.AddDate(0, 0, 1)}},
	}
	title, body, ok := ReminderText(types, day)
	if !ok || title != "Tomorrow: Yellow bag, Paper" || body == "" {
		t.Fatalf("%q %q %v", title, body, ok)
	}
	if _, _, ok := ReminderText(types, day.AddDate(0, 0, 2)); ok {
		t.Fatal("no message without a collection")
	}
}

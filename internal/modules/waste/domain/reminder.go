package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"waylight/internal/shared/kernel"
)

// DefaultReminderMinute is 19:00, the time a new reminder starts with (#866).
const DefaultReminderMinute = 19 * 60

// Reminder is the message the evening before a collection (#866). It is off
// until someone turns it on; Minute is the time of day in minutes.
type Reminder struct {
	Enabled bool
	Minute  int
}

// DefaultReminder is off at 19:00.
func DefaultReminder() Reminder { return Reminder{Minute: DefaultReminderMinute} }

// NewReminder validates a reminder; at is "HH:MM".
func NewReminder(enabled bool, at string) (Reminder, error) {
	t, err := time.Parse("15:04", at)
	if err != nil || len(at) != len("15:04") {
		return Reminder{}, kernel.NewValidationError("time", "Enter a time like 19:00.")
	}
	return Reminder{Enabled: enabled, Minute: t.Hour()*60 + t.Minute()}, nil
}

// At returns the time of day as "HH:MM".
func (r Reminder) At() string { return fmt.Sprintf("%02d:%02d", r.Minute/60, r.Minute%60) }

// Due returns the collection day (tomorrow, as Date) a reminder is due for
// at now (in the user's time zone): once the time has passed, and not yet
// sent for that day. A server that was down at the time sends it later
// that evening.
func (r Reminder) Due(now, sentFor time.Time) (time.Time, bool) {
	if !r.Enabled {
		return time.Time{}, false
	}
	y, m, d := now.Date()
	fire := time.Date(y, m, d, r.Minute/60, r.Minute%60, 0, 0, now.Location())
	tomorrow := Date(now.AddDate(0, 0, 1))
	if now.Before(fire) || !sentFor.Before(tomorrow) {
		return time.Time{}, false
	}
	return tomorrow, true
}

// ReminderText is the message for the collections on day, taken from the
// upcoming dates per type: e.g. "Tomorrow: Yellow bag, Paper". ok is false
// when nothing is collected that day. Push texts are English like the other
// server notifications.
func ReminderText(types []Upcoming, day time.Time) (title, body string, ok bool) {
	var labels []string
	for _, u := range types {
		if slices.ContainsFunc(u.Dates, day.Equal) && !slices.Contains(labels, u.Style.Label) {
			labels = append(labels, u.Style.Label)
		}
	}
	if len(labels) == 0 {
		return "", "", false
	}
	return "Tomorrow: " + strings.Join(labels, ", "), "Waste collection tomorrow. Put the bins out tonight.", true
}

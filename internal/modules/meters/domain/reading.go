package domain

import (
	"math"
	"time"

	"waylight/internal/shared/kernel"
)

// Reading is the counter value on one day (midnight UTC).
type Reading struct {
	Day   time.Time
	Value float64
}

// ReadingInput is what a person enters for a reading.
type ReadingInput struct {
	Day   string // YYYY-MM-DD
	Value float64
}

// NewReading validates a reading; today is the viewer's day, so a reading
// cannot lie in the future.
func NewReading(in ReadingInput, today time.Time) (Reading, error) {
	day, err := ParseDay("day", in.Day)
	if err != nil {
		return Reading{}, err
	}
	if day.Year() < MinYear {
		return Reading{}, kernel.NewValidationError("day", "Enter a date like 2026-12-24.")
	}
	if day.After(dateOf(today)) {
		return Reading{}, kernel.NewValidationError("day", "The day cannot be in the future.")
	}
	if math.IsNaN(in.Value) || in.Value < 0 || in.Value > MaxValue {
		return Reading{}, kernel.NewValidationError("value", "Enter the counter value, zero or more.")
	}
	return Reading{Day: day, Value: in.Value}, nil
}

// CheckOrder refuses a reading that is lower than an earlier one or higher
// than a later one, since a counter only counts up. A reading on a day that
// already has one replaces it and is checked against the other days.
func CheckOrder(existing []Reading, r Reading) error {
	for _, e := range existing {
		switch {
		case e.Day.Before(r.Day) && e.Value > r.Value:
			return kernel.NewValidationError("value", "The reading cannot be lower than an earlier one.")
		case e.Day.After(r.Day) && e.Value < r.Value:
			return kernel.NewValidationError("value", "The reading cannot be higher than a later one.")
		}
	}
	return nil
}

func dateOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

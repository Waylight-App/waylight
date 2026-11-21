package usecase

import (
	"time"

	"waylight/internal/shared/kernel"
)

// Layouts of event times in a compact state: numeric, so they read the
// same in every language.
const (
	compactClock = "15:04"
	compactDate  = "2006-01-02"
)

// CompactState shows the next event that has not ended yet (#594, #597):
// its title, and when it starts relative to today. An event is neither
// good nor bad, so the level stays none. Times are in loc.
func CompactState(agenda AgendaResponse, now time.Time, loc *time.Location) kernel.CompactState {
	state := kernel.CompactState{Level: kernel.CompactNone, UpdatedAt: now}
	if !agenda.UpdatedAt.IsZero() {
		state.UpdatedAt = agenda.UpdatedAt
	}
	var next *EventView
	for i := range agenda.Events {
		e := &agenda.Events[i]
		if e.End.After(now) && (next == nil || e.Start.Before(next.Start)) {
			next = e
		}
	}
	if next == nil {
		state.Title = kernel.NewText("No upcoming events")
		state.Empty = true
		return state
	}
	state.Title = kernel.NewTextf("{title}", "title", next.Summary)
	start := next.Start.In(loc)
	today := now.In(loc)
	day := dayDistance(today, start)
	clock := start.Format(compactClock)
	switch {
	case !start.After(today):
		state.Detail = kernel.NewText("Now")
	case next.AllDay && day == 0:
		state.Detail = kernel.NewText("Today")
	case next.AllDay && day == 1:
		state.Detail = kernel.NewText("Tomorrow")
	case next.AllDay:
		state.Detail = kernel.NewTextf("On {date}", "date", start.Format(compactDate))
	case day == 0:
		state.Detail = kernel.NewTextf("Today at {time}", "time", clock)
	case day == 1:
		state.Detail = kernel.NewTextf("Tomorrow at {time}", "time", clock)
	default:
		state.Detail = kernel.NewTextf("On {date} at {time}", "date", start.Format(compactDate), "time", clock)
	}
	return state
}

// dayDistance counts calendar days from a to b in a's location.
func dayDistance(a, b time.Time) int {
	da := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
	db := time.Date(b.Year(), b.Month(), b.Day(), 0, 0, 0, 0, time.UTC)
	return int(db.Sub(da).Hours() / 24)
}

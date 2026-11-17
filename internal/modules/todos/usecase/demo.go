package usecase

import (
	"context"
	"fmt"
	"time"
)

// demoMarker is the first to-do SeedDemo adds; when it exists the demo
// to-dos are already there.
const demoMarker = "Book the car inspection"

// SeedDemo adds everyday to-dos with due dates to the public demo, next to
// the sample to-dos, so the list, the briefing and the search have
// something to show.
func SeedDemo(ctx context.Context, d Deps) error {
	existing, err := d.Repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list todos: %w", err)
	}
	for _, t := range existing {
		if t.Title() == demoMarker {
			return nil
		}
	}
	today := d.Clock.Now()
	samples := []struct {
		title string
		due   time.Time
		done  bool
	}{
		{demoMarker, today, false},
		{"Renew the TLS certificate of the NAS", today.AddDate(0, 0, -1), false},
		{"Order a present for Anna", today.AddDate(0, 0, 2), false},
		{"Water the plants", today, true},
		{"Update Jellyfin to the new release", today.AddDate(0, 0, 5), false},
		{"Pay the electricity bill", today.AddDate(0, 0, 7), false},
	}
	create, update := NewCreateTodo(d), NewUpdateTodo(d)
	for _, s := range samples {
		t, err := create.Execute(ctx, TodoInput{Title: s.title, Due: s.due})
		if err != nil {
			return err
		}
		if s.done {
			if _, err := update.Execute(ctx, t.ID, TodoInput{Title: s.title, Due: s.due}, true); err != nil {
				return err
			}
		}
	}
	return nil
}

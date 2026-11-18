package usecase

import (
	"context"
	"fmt"

	"waylight/internal/modules/habits/domain"
)

// demoHabit is a habit of the public demo with the days it was done, counted
// back from today (0 = today).
type demoHabit struct {
	name, emoji string
	days        []int
	skip        func(daysAgo int) bool
}

// SeedDemo adds habits with four weeks of history to the public demo, so
// the widget shows streaks and a filled week.
func SeedDemo(ctx context.Context, d Deps) error {
	existing, err := d.Repo.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	started := now.AddDate(0, 0, -30)
	habits := []demoHabit{
		{name: "Drink water", emoji: "💧", skip: func(n int) bool { return n == 9 }},
		{name: "Read 20 minutes", emoji: "📚", skip: func(n int) bool { return n%6 == 5 }},
		{name: "Workout", emoji: "🏃", days: []int{1, 3, 5}, skip: func(n int) bool { return n == 14 }},
		{name: "No phone after 10 pm", emoji: "🌙", skip: func(n int) bool { return n%4 == 2 || n == 0 }},
	}
	for i, dh := range habits {
		s, err := domain.NewSchedule(dh.days)
		if err != nil {
			return err
		}
		h, err := domain.New(d.IDs.NewID(), dh.name, dh.emoji, s, i, started)
		if err != nil {
			return err
		}
		if err := d.Repo.Save(ctx, h); err != nil {
			return fmt.Errorf("save habit: %w", err)
		}
		for n := 1; n <= 28; n++ {
			day := domain.DayOf(now.AddDate(0, 0, -n))
			if !s.Due(day) || dh.skip(n) {
				continue
			}
			if err := d.Repo.SetCompletion(ctx, h.ID, day, true); err != nil {
				return fmt.Errorf("save completion: %w", err)
			}
		}
	}
	return nil
}

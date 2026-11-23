package usecase

import (
	"context"
	"time"

	"waylight/internal/modules/countdown/domain"
)

// SeedDemo adds a few countdowns to the public demo, dated from today.
func SeedDemo(ctx context.Context, d Deps) error {
	existing, err := d.Repo.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	day := func(days int) string { return now.AddDate(0, 0, days).Format(time.DateOnly) }
	uc := NewCountdowns(d)
	for _, in := range []domain.Input{
		{Title: "Car inspection", Date: day(4), Time: "08:30"},
		{Title: "Anna's birthday party", Date: day(9), Time: "18:00"},
		{Title: "Concert at the Elbphilharmonie", Date: day(23), Time: "20:00"},
		{Title: "Summer holiday on Norderney", Date: day(46)},
	} {
		if _, err := uc.Create(ctx, in); err != nil {
			return err
		}
	}
	return nil
}

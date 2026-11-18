package usecase

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/modules/focus/domain"
)

// SeedDemo adds a week of finished focus sessions to the public demo, so
// the widget's week chart is filled. The sessions are built directly: the
// regular use case only accepts sessions of the last day.
func SeedDemo(ctx context.Context, d Deps) error {
	existing, err := d.Repo.Since(ctx, time.Time{})
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	perDay := []int{0, 3, 4, 2, 5, 1, 3} // index = days ago; sessions of 25 minutes
	for daysAgo, count := range perDay {
		end := now.AddDate(0, 0, -daysAgo).Truncate(time.Hour)
		for i := range count {
			s := domain.Session{ID: d.IDs.NewID(), Minutes: 25, EndedAt: end.Add(-time.Duration(i) * 35 * time.Minute)}
			if err := d.Repo.Save(ctx, s); err != nil {
				return fmt.Errorf("save session: %w", err)
			}
		}
	}
	return nil
}

package usecase

import (
	"context"
	"time"

	"waylight/internal/modules/board/domain"
)

// SeedDemo pins a few notes on the board of the public demo, once.
func SeedDemo(ctx context.Context, d Deps) error {
	b := NewBoard(d)
	existing, err := b.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	day := func(n int) string { return d.Clock.Now().AddDate(0, 0, n).Format(time.DateOnly) }
	for _, in := range []domain.Input{
		{Text: "Wi-Fi for guests: Waylight-Guest, the password is on the fridge.", Pinned: true},
		{Text: "Mia has swimming on Thursday, bring the blue towel.", ExpiresOn: day(3)},
		{Text: "Parcel for the neighbours is in the hallway.", ExpiresOn: day(1)},
		{Text: "Grandma visits on Sunday, lunch at 12."},
	} {
		if _, err := b.Create(ctx, in, ""); err != nil {
			return err
		}
	}
	return nil
}

package usecase

import (
	"context"
	"time"
)

// SeedDemo leaves one text in the drop zone of the public demo, as if it
// had just been sent from a phone.
func SeedDemo(ctx context.Context, d Drop) error {
	existing, err := d.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	_, err = d.AddText(ctx, "Guest Wi-Fi: Waylight-Guest · password: lamp-light-2026", 24*time.Hour)
	return err
}

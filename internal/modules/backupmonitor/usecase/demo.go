package usecase

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/modules/backupmonitor/domain"
)

// demoSource is a monitored job of the public demo; the times are relative
// to now, zero means never.
type demoSource struct {
	name        string
	schedule    domain.Schedule
	lastSuccess time.Duration
	lastFailure time.Duration
	duration    time.Duration
	message     string
}

// SeedDemo adds backup jobs and heartbeat checks to the public demo: most
// are fine, one backup is overdue and one check failed, so every state of
// the widget is visible.
func SeedDemo(ctx context.Context, d Deps) error {
	existing, err := d.Repo.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	now := d.Clock.Now()
	day := 24 * time.Hour
	// The nightly backup last finished at 03:17, today or yesterday.
	nightly := time.Date(now.Year(), now.Month(), now.Day(), 3, 17, 0, 0, now.Location())
	if nightly.After(now) {
		nightly = nightly.Add(-day)
	}
	sources := []demoSource{
		{name: "restic → Hetzner Storage Box", schedule: domain.Schedule{Kind: "backup", Cron: "0 3 * * *", Grace: 2 * time.Hour},
			lastSuccess: now.Sub(nightly), duration: 23 * time.Minute, message: "snapshot 4f2a91c saved, 1.2 GiB added"},
		{name: "Paperless export", schedule: domain.Schedule{Kind: "backup", MaxAge: day, Grace: time.Hour},
			lastSuccess: 5 * time.Hour, duration: 2 * time.Minute},
		{name: "Photos → USB disk", schedule: domain.Schedule{Kind: "backup", MaxAge: 7 * day, Grace: day},
			lastSuccess: 10 * day, duration: 48 * time.Minute, message: "disk not mounted"},
		{name: "Home Assistant snapshot", schedule: domain.Schedule{Kind: "backup", MaxAge: day, Grace: 2 * time.Hour},
			lastSuccess: 9 * time.Hour, duration: 40 * time.Second},
		{name: "Certificate renewal", schedule: domain.Schedule{Kind: "check", MaxAge: 12 * time.Hour, Grace: time.Hour},
			lastSuccess: 3 * time.Hour},
		{name: "Offsite sync (rclone)", schedule: domain.Schedule{Kind: "check", MaxAge: 6 * time.Hour, Grace: time.Hour},
			lastSuccess: 7 * time.Hour, lastFailure: 40 * time.Minute, message: "Connection reset by peer"},
	}
	for _, ds := range sources {
		token, err := d.Tokens.NewToken()
		if err != nil {
			return fmt.Errorf("generate token: %w", err)
		}
		s, err := domain.NewSource(d.IDs.NewID(), ds.name, ds.schedule, domain.HashToken(token), now.Add(-60*day))
		if err != nil {
			return err
		}
		if ds.lastSuccess > 0 {
			s.LastSuccessAt = now.Add(-ds.lastSuccess)
			s.LastStartAt = s.LastSuccessAt.Add(-ds.duration)
			s.LastDuration = ds.duration
		}
		if ds.lastFailure > 0 {
			s.LastFailureAt = now.Add(-ds.lastFailure)
		}
		s.LastMessage = ds.message
		if err := d.Repo.Save(ctx, s); err != nil {
			return fmt.Errorf("save source: %w", err)
		}
	}
	return nil
}

package usecase

import (
	"context"
	"fmt"
	"time"

	"waylight/internal/modules/timeline/domain"
	"waylight/internal/shared/kernel"
)

// SeedDemo fills an empty timeline of the public demo with a plausible day
// of a homelab: what the regular sources would have recorded. It does
// nothing when the timeline already has events.
func SeedDemo(ctx context.Context, repo Repository, clock kernel.Clock, ids kernel.IDGenerator) error {
	existing, err := repo.List(ctx, Filter{Limit: 1})
	if err != nil || len(existing) > 0 {
		return err
	}
	now := clock.Now()
	events := []struct {
		ago      time.Duration
		source   string
		severity domain.Severity
		subject  string
		message  string
	}{
		{25 * time.Minute, "container", domain.SeverityInfo, "jellyfin", "jellyfin was restarted after an image update"},
		{70 * time.Minute, "disk", domain.SeverityWarning, "/", "Disk / is 85 % full"},
		{2 * time.Hour, "parcel", domain.SeverityInfo, "coffee", "Parcel \"Coffee beans\" was delivered"},
		{3 * time.Hour, "service", domain.SeverityInfo, "nextcloud", "Nextcloud is OK again"},
		{3*time.Hour + 6*time.Minute, "service", domain.SeverityCritical, "nextcloud", "Nextcloud is not reachable"},
		{6 * time.Hour, "backup", domain.SeverityCritical, "restic-offsite", "Backup restic-offsite failed: repository locked"},
		{7 * time.Hour, "sports", domain.SeverityInfo, "bundesliga", "Werder Bremen 2 : 1 VfL Wolfsburg"},
		{9 * time.Hour, "tls", domain.SeverityWarning, "cloud", "The certificate of cloud.example.org expires in 12 days"},
		{20 * time.Hour, "backup", domain.SeverityInfo, "paperless-export", "Backup paperless-export finished"},
		{29 * time.Hour, "diskhealth", domain.SeverityInfo, "sda", "Disk sda passed its S.M.A.R.T. self-test"},
	}
	for _, e := range events {
		ev := domain.NewEvent(ids.NewID(), now.Add(-e.ago), e.source, e.severity, e.subject, e.message)
		if err := repo.Add(ctx, ev); err != nil {
			return fmt.Errorf("add demo event: %w", err)
		}
	}
	return nil
}

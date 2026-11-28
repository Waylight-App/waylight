package gateway

import (
	"context"
	"testing"
	"time"

	"waylight/internal/modules/waste/domain"
	"waylight/internal/platform/database/dbtest"
)

func TestReminderRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewReminderRepository(dbtest.Open(t, "waste", Migrations()))
	r, sent, err := repo.LoadReminder(ctx)
	if err != nil || r != domain.DefaultReminder() || !sent.IsZero() {
		t.Fatalf("default: %+v %v %v", r, sent, err)
	}
	day := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	if err := repo.MarkReminderSent(ctx, day); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveReminder(ctx, domain.Reminder{Enabled: true, Minute: 18 * 60}); err != nil {
		t.Fatal(err)
	}
	r, sent, err = repo.LoadReminder(ctx)
	if err != nil || !r.Enabled || r.Minute != 18*60 || !sent.Equal(day) {
		t.Fatalf("saving keeps the sent day: %+v %v %v", r, sent, err)
	}
}

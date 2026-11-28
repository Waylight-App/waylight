package usecase

import (
	"context"
	"testing"
	"time"

	"waylight/internal/modules/waste/domain"
	"waylight/internal/shared/kernel"
)

type memReminders struct {
	r       *domain.Reminder
	sentFor time.Time
}

func (m *memReminders) LoadReminder(context.Context) (domain.Reminder, time.Time, error) {
	if m.r == nil {
		return domain.DefaultReminder(), m.sentFor, nil
	}
	return *m.r, m.sentFor, nil
}
func (m *memReminders) SaveReminder(_ context.Context, r domain.Reminder) error { m.r = &r; return nil }
func (m *memReminders) MarkReminderSent(_ context.Context, day time.Time) error {
	m.sentFor = day
	return nil
}

type alertLog []kernel.Alert

func (l *alertLog) Alert(_ context.Context, a kernel.Alert) { *l = append(*l, a) }

func TestReminders(t *testing.T) {
	ctx := context.Background()
	settings := &memRepo{}
	waste := NewWaste(settings, &fakeSource{}, fakeParser{}, clk{})
	repo, alerts := &memReminders{}, &alertLog{}
	uc := NewReminders(repo, waste, alerts, clk{}, time.UTC)

	if r, err := uc.Reminder(ctx); err != nil || r.Enabled || r.At() != "19:00" {
		t.Fatalf("default: %+v %v", r, err)
	}
	if err := uc.Send(ctx); err != nil || len(*alerts) != 0 {
		t.Fatalf("off by default: %v %v", *alerts, err)
	}
	if _, err := uc.SaveReminder(ctx, true, "19:00"); err != nil {
		t.Fatal(err)
	}
	if err := uc.Send(ctx); err != nil || len(*alerts) != 0 {
		t.Fatalf("no calendar, no message: %v %v", *alerts, err)
	}
	if _, err := waste.SaveSettings(ctx, "https://example.org/a.ics", nil); err != nil {
		t.Fatal(err)
	}
	// The clock reads 11 Nov 19:00; the organic bin goes out on 12 Nov.
	if err := uc.Send(ctx); err != nil || len(*alerts) != 1 {
		t.Fatalf("reminder: %v %v", *alerts, err)
	}
	if a := (*alerts)[0]; a.Kind != "reminder" || a.Title != "Tomorrow: Biotonne" {
		t.Fatalf("%+v", a)
	}
	if err := uc.Send(ctx); err != nil || len(*alerts) != 1 {
		t.Fatalf("sent once per day: %v %v", *alerts, err)
	}
	if _, err := uc.SaveReminder(ctx, true, "7pm"); err == nil {
		t.Fatal("invalid time must be rejected")
	}
}

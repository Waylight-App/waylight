package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/waste/domain"
	"waylight/internal/shared/kernel"
)

// Reminders is the input port of the evening-before reminder (#866).
type Reminders interface {
	Reminder(ctx context.Context) (domain.Reminder, error)
	SaveReminder(ctx context.Context, enabled bool, at string) (domain.Reminder, error)
	// Send is run by the background job every minute; it sends the
	// reminder once per collection day through the notification channels.
	Send(ctx context.Context) error
}

// ReminderRepository persists the reminder and the day it was last sent
// for (output port). Without a stored row it returns domain.DefaultReminder.
type ReminderRepository interface {
	LoadReminder(ctx context.Context) (r domain.Reminder, sentFor time.Time, err error)
	SaveReminder(ctx context.Context, r domain.Reminder) error
	MarkReminderSent(ctx context.Context, day time.Time) error
}

type reminders struct {
	repo   ReminderRepository
	waste  Waste
	alerts kernel.AlertSink
	clock  kernel.Clock
	loc    *time.Location
}

// NewReminders returns the reminder use cases; the dates come from waste,
// the message goes to alerts (the existing notification channels).
func NewReminders(repo ReminderRepository, waste Waste, alerts kernel.AlertSink, clock kernel.Clock, loc *time.Location) Reminders {
	return &reminders{repo: repo, waste: waste, alerts: alerts, clock: clock, loc: loc}
}

// Reminder implements Reminders.
func (uc *reminders) Reminder(ctx context.Context) (domain.Reminder, error) {
	r, _, err := uc.repo.LoadReminder(ctx)
	if err != nil {
		return domain.Reminder{}, fmt.Errorf("load reminder: %w", err)
	}
	return r, nil
}

// SaveReminder implements Reminders.
func (uc *reminders) SaveReminder(ctx context.Context, enabled bool, at string) (domain.Reminder, error) {
	r, err := domain.NewReminder(enabled, at)
	if err != nil {
		return domain.Reminder{}, err
	}
	if err := uc.repo.SaveReminder(ctx, r); err != nil {
		return domain.Reminder{}, fmt.Errorf("save reminder: %w", err)
	}
	return r, nil
}

// Send implements Reminders.
func (uc *reminders) Send(ctx context.Context) error {
	r, sentFor, err := uc.repo.LoadReminder(ctx)
	if err != nil {
		return fmt.Errorf("load reminder: %w", err)
	}
	day, due := r.Due(uc.clock.Now().In(uc.loc), sentFor)
	if !due {
		return nil
	}
	o, err := uc.waste.Upcoming(ctx)
	if errors.Is(err, domain.ErrNotConfigured) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("upcoming collections: %w", err)
	}
	title, body, ok := domain.ReminderText(o.Types, day)
	if !ok {
		return nil
	}
	// Marked first: a failed write must not send the message every minute.
	if err := uc.repo.MarkReminderSent(ctx, day); err != nil {
		return fmt.Errorf("mark reminder sent: %w", err)
	}
	uc.alerts.Alert(ctx, kernel.Alert{Kind: "reminder", Title: title, Body: body})
	return nil
}

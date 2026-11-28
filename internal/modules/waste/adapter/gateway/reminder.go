package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/waste/domain"
)

// ReminderRepository stores the evening-before reminder (#866).
type ReminderRepository struct{ db *sql.DB }

// NewReminderRepository returns a repository backed by db.
func NewReminderRepository(db *sql.DB) *ReminderRepository { return &ReminderRepository{db: db} }

// LoadReminder implements usecase.ReminderRepository.
func (r *ReminderRepository) LoadReminder(ctx context.Context) (domain.Reminder, time.Time, error) {
	var rem domain.Reminder
	var sent string
	err := r.db.QueryRowContext(ctx, `SELECT enabled, minute, sent_for FROM waste_reminder WHERE id = 1`).
		Scan(&rem.Enabled, &rem.Minute, &sent)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.DefaultReminder(), time.Time{}, nil
	}
	if err != nil {
		return domain.Reminder{}, time.Time{}, fmt.Errorf("select reminder: %w", err)
	}
	sentFor, _ := time.Parse(time.DateOnly, sent) // empty until the first reminder
	return rem, sentFor, nil
}

// SaveReminder implements usecase.ReminderRepository; the sent day stays.
func (r *ReminderRepository) SaveReminder(ctx context.Context, rem domain.Reminder) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO waste_reminder (id, enabled, minute) VALUES (1, ?, ?)
		ON CONFLICT (id) DO UPDATE SET enabled = excluded.enabled, minute = excluded.minute`, rem.Enabled, rem.Minute)
	if err != nil {
		return fmt.Errorf("upsert reminder: %w", err)
	}
	return nil
}

// MarkReminderSent implements usecase.ReminderRepository.
func (r *ReminderRepository) MarkReminderSent(ctx context.Context, day time.Time) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO waste_reminder (id, sent_for) VALUES (1, ?)
		ON CONFLICT (id) DO UPDATE SET sent_for = excluded.sent_for`, day.Format(time.DateOnly))
	if err != nil {
		return fmt.Errorf("mark reminder sent: %w", err)
	}
	return nil
}

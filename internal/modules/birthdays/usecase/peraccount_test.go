package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/birthdays/domain"
	"trisse-home/internal/shared/kernel"
)

// TestBirthdayReminderIsAdminOnly: reminders go to the instance's channels,
// so members cannot switch them on (ADR 0018).
func TestBirthdayReminderIsAdminOnly(t *testing.T) {
	uc := newTestBirthdays(t)
	member := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if _, err := uc.SaveSettings(member, SettingsInput{RemindEnabled: true, RemindAt: "08:00"}); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatalf("member switched on reminders: %v", err)
	}
	if _, err := uc.SaveSettings(member, SettingsInput{RemindAt: "08:00"}); err != nil {
		t.Fatalf("member saving without reminder: %v", err)
	}
}

func newTestBirthdays(t *testing.T) Birthdays {
	t.Helper()
	return New(Deps{Repo: &repo{m: map[string]domain.Birthday{}}, Book: &book{}, IDs: &ids{},
		Clock: &clk{t: time.Date(2026, 11, 14, 7, 0, 0, 0, time.UTC)}, Location: time.UTC, Alerts: &alerts{}})
}

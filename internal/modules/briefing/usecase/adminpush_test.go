package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/shared/kernel"
)

// TestPushIsAdminOnly: the scheduled briefing goes to the instance's
// channels, so members can arrange their briefing but not push it (ADR 0018).
func TestPushIsAdminOnly(t *testing.T) {
	uc := NewManageSettings(Deps{Repo: &memRepo{}})
	member := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if _, err := uc.Save(member, SettingsInput{PushOn: true, PushAt: "07:00"}); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatalf("member switched on the push: %v", err)
	}
	if _, err := uc.Save(member, SettingsInput{PushAt: "07:00", Lat: 48.1, Lon: 11.6}); err != nil {
		t.Fatalf("member saving without push: %v", err)
	}
	if err := NewSendBriefing(Deps{Repo: &memRepo{}, Publisher: &publisher{}}).SendNow(member); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatalf("member sent the briefing to the instance channels: %v", err)
	}
}

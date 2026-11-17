package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/bookmarks/domain"
	"waylight/internal/shared/kernel"
)

type memSettings struct{ s domain.Settings }

func (m *memSettings) LoadSettings(context.Context) (domain.Settings, error) { return m.s, nil }
func (m *memSettings) SaveSettings(_ context.Context, s domain.Settings) error {
	m.s = s
	return nil
}

// TestOutboundSettingsAreAdminOnly: fetching icons and checking links is an
// instance decision, so members cannot switch it (ADR 0018).
func TestOutboundSettingsAreAdminOnly(t *testing.T) {
	store := &memSettings{}
	uc := NewSaveSettings(Deps{Settings: store})
	member := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if _, err := uc.Execute(member, SettingsView{Enrich: true}); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatalf("member switched outbound features: %v", err)
	}
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if _, err := uc.Execute(admin, SettingsView{Enrich: true}); err != nil || !store.s.Enrich {
		t.Fatalf("admin: %v %+v", err, store.s)
	}
}

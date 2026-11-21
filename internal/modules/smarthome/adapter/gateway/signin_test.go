package gateway

import (
	"context"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/smarthome/domain"
	"waylight/internal/platform/database/dbtest"
)

// TestSignedInSettings: the refresh token of a sign-in (#654) is stored
// encrypted and comes back with client ID and expiry.
func TestSignedInSettings(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "smarthome", Migrations())
	repo := NewSettingsRepository(db, prefixCipher{})
	s := domain.Settings{URL: "http://ha:8123", Token: "access", Refresh: "refresh-secret",
		ClientID: "https://home.example/", Expiry: time.Unix(1_700_000_000, 0)}
	if err := repo.Save(ctx, s); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT refresh_encrypted FROM smarthome_settings`).Scan(&stored)
	if stored == "" || strings.Contains(stored, "refresh-secret") {
		t.Fatalf("refresh token stored in plain text: %q", stored)
	}
	got, err := repo.Load(ctx)
	if err != nil || got.Refresh != s.Refresh || got.ClientID != s.ClientID || !got.Expiry.Equal(s.Expiry) || got.Token != s.Token {
		t.Fatalf("round trip: %+v %v", got, err)
	}
}

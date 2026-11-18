package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/platform/database/dbtest"
)

func TestPairingCodesAreSingleUseAndExpire(t *testing.T) {
	ctx := context.Background()
	repo := NewPairingCodeRepository(dbtest.Open(t, "auth", Migrations()))
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	for _, raw := range []string{"used", "late"} {
		c, _ := domain.NewPairingCode(domain.HashToken(raw), "owner", now)
		if err := repo.Create(ctx, c, now); err != nil {
			t.Fatal(err)
		}
	}
	if id, err := repo.Take(ctx, domain.HashToken("used"), now.Add(time.Minute)); err != nil || id != "owner" {
		t.Fatalf("take: %q, %v", id, err)
	}
	if _, err := repo.Take(ctx, domain.HashToken("used"), now.Add(time.Minute)); !errors.Is(err, domain.ErrInvalidPairingCode) {
		t.Fatalf("second take: %v", err)
	}
	if _, err := repo.Take(ctx, domain.HashToken("late"), now.Add(domain.PairingCodeTTL)); !errors.Is(err, domain.ErrInvalidPairingCode) {
		t.Fatalf("expired take: %v", err)
	}
	if _, err := repo.Take(ctx, domain.HashToken("unknown"), now); !errors.Is(err, domain.ErrInvalidPairingCode) {
		t.Fatalf("unknown take: %v", err)
	}
}

func TestSessionKeepsDeviceName(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "auth", Migrations())
	accounts := NewAccountRepository(db, prefixCipher{})
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	a, _ := domain.NewAccount("hash", now)
	if err := accounts.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	sessions := NewSessionRepository(db)
	s, _ := domain.NewAppSession(a.ID(), domain.HashToken("wa_t"), "p1", "csrf", domain.Client{}, "Pixel 9", now)
	if err := sessions.Create(ctx, s); err != nil {
		t.Fatal(err)
	}
	got, err := sessions.Find(ctx, domain.HashToken("wa_t"))
	if err != nil || got.Device() != "Pixel 9" || !got.IsApp() {
		t.Fatalf("got %+v, %v", got, err)
	}
}

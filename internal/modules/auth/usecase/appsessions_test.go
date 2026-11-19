package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/auth/domain"
	"waylight/internal/shared/kernel"
)

// fakePairing keeps pairing codes in memory with the repository's rules.
type fakePairing struct {
	codes map[domain.TokenHash]domain.PairingCode
}

func (f *fakePairing) Create(_ context.Context, c domain.PairingCode, _ time.Time) error {
	f.codes[c.Hash] = c
	return nil
}

func (f *fakePairing) Take(_ context.Context, h domain.TokenHash, now time.Time) (domain.PairingCode, error) {
	c, ok := f.codes[h]
	delete(f.codes, h)
	if !ok || !now.Before(c.ExpiresAt) {
		return domain.PairingCode{}, domain.ErrInvalidPairingCode
	}
	return c, nil
}

func newAppSessionsEnv() (*env, AppSessions) {
	e := newEnv().withAccount()
	return e, NewAppSessions(e.deps, &fakePairing{codes: map[domain.TokenHash]domain.PairingCode{}})
}

func TestAppSessionWithPassword(t *testing.T) {
	e, uc := newAppSessionsEnv()
	res, err := uc.Create(context.Background(), AppSessionRequest{Login: LoginRequest{Password: goodPassword, ClientKey: "ip"}, Device: "Pixel 9"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(res.Token, domain.AppSessionTokenPrefix) || res.Device != "Pixel 9" ||
		!res.ExpiresAt.Equal(e.clock.now.Add(domain.AppSessionTTL)) {
		t.Fatalf("unexpected response %+v", res)
	}
	s, err := e.sessions.Find(context.Background(), domain.HashToken(res.Token))
	if err != nil || !s.IsApp() {
		t.Fatalf("session not stored as app session: %+v, %v", s, err)
	}
}

func TestAppSessionRejectsWrongPasswordAndMissingDevice(t *testing.T) {
	_, uc := newAppSessionsEnv()
	ctx := context.Background()
	if _, err := uc.Create(ctx, AppSessionRequest{Login: LoginRequest{Password: "wrong", ClientKey: "ip"}, Device: "Pixel"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	if _, err := uc.Create(ctx, AppSessionRequest{Login: LoginRequest{Password: goodPassword, ClientKey: "ip"}}); err == nil {
		t.Fatal("missing device accepted")
	}
}

func TestPairingCodeIsRedeemedOnce(t *testing.T) {
	e, uc := newAppSessionsEnv()
	owner := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: domain.OwnerID, Role: string(domain.RoleOwner)})
	code, err := uc.NewPairingCode(owner)
	if err != nil {
		t.Fatal(err)
	}
	if !code.ExpiresAt.Equal(e.clock.now.Add(domain.PairingCodeTTL)) {
		t.Fatalf("expiry %v", code.ExpiresAt)
	}
	req := AppSessionRequest{PairingCode: code.Code, Device: "iPhone", Login: LoginRequest{ClientKey: "ip"}}
	res, err := uc.Create(context.Background(), req)
	if err != nil || !strings.HasPrefix(res.Token, domain.AppSessionTokenPrefix) {
		t.Fatalf("pair: %+v, %v", res, err)
	}
	if _, err := uc.Create(context.Background(), req); !errors.Is(err, domain.ErrInvalidPairingCode) {
		t.Fatalf("second redemption: %v", err)
	}
	if e.limiter.failures["ip"] != 1 {
		t.Fatalf("failed redemption not counted: %d", e.limiter.failures["ip"])
	}
}

func TestKioskCannotCreatePairingCode(t *testing.T) {
	_, uc := newAppSessionsEnv()
	kiosk := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "k1", Role: string(domain.RoleKiosk)})
	if _, err := uc.NewPairingCode(kiosk); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("kiosk got a code: %v", err)
	}
}

func TestAppSessionsRefreshWithTheirOwnLifetime(t *testing.T) {
	e, uc := newAppSessionsEnv()
	res, err := uc.Create(context.Background(), AppSessionRequest{Login: LoginRequest{Password: goodPassword, ClientKey: "ip"}, Device: "Pixel 9"})
	if err != nil {
		t.Fatal(err)
	}
	auth := NewAuthenticate(e.sessions, e.accounts, nil, nil, e.clock, e.deps.TTL)
	e.clock.now = e.clock.now.Add(48 * time.Hour) // beyond the browser TTL of the env
	got, err := auth.Execute(context.Background(), AuthenticateRequest{SessionToken: res.Token})
	if err != nil {
		t.Fatal(err)
	}
	if !got.ExpiresAt.Equal(e.clock.now.Add(domain.AppSessionTTL)) {
		t.Fatalf("refreshed to %v", got.ExpiresAt)
	}
}

func TestHandoverCodeNeedsTheVerifier(t *testing.T) {
	e, uc := newAppSessionsEnv()
	owner := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: domain.OwnerID, Role: string(domain.RoleOwner)})
	verifier := strings.Repeat("v", 43)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	if _, err := uc.NewHandoverCode(owner, "too-short"); err == nil {
		t.Fatal("invalid challenge accepted")
	}
	stolen, _ := uc.NewHandoverCode(owner, challenge)
	if !stolen.ExpiresAt.Equal(e.clock.now.Add(domain.HandoverCodeTTL)) {
		t.Fatalf("expiry %v", stolen.ExpiresAt)
	}
	req := AppSessionRequest{PairingCode: stolen.Code, Verifier: "wrong", Device: "Pixel", Login: LoginRequest{ClientKey: "ip"}}
	if _, err := uc.Create(context.Background(), req); !errors.Is(err, domain.ErrInvalidPairingCode) {
		t.Fatalf("wrong verifier: %v", err)
	}
	code, _ := uc.NewHandoverCode(owner, challenge)
	req.PairingCode, req.Verifier = code.Code, verifier
	if res, err := uc.Create(context.Background(), req); err != nil || !strings.HasPrefix(res.Token, domain.AppSessionTokenPrefix) {
		t.Fatalf("handover: %+v, %v", res, err)
	}
}

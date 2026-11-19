package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/shared/kernel"
)

type memDevices map[string]domain.AppDevice

func (m memDevices) List(context.Context) ([]domain.AppDevice, error) {
	out := make([]domain.AppDevice, 0, len(m))
	for _, d := range m {
		out = append(out, d)
	}
	return out, nil
}
func (m memDevices) Save(_ context.Context, d domain.AppDevice) error { m[d.Session] = d; return nil }
func (m memDevices) Delete(_ context.Context, s string) error         { delete(m, s); return nil }

type fakeRelay struct {
	pushed []string
	gone   map[string]bool
}

func (f *fakeRelay) Push(_ context.Context, d domain.AppDevice, _ domain.Message) error {
	if f.gone[d.Session] {
		return ErrDeviceGone
	}
	f.pushed = append(f.pushed, d.Session)
	return nil
}

type activeSessions map[string]bool

func (a activeSessions) AppSessionActive(_ context.Context, _, id string) (bool, error) {
	return a[id], nil
}

func appCtx(role, session string) context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "acc", Role: role, AppSession: session})
}

func TestAppDevicesRegister(t *testing.T) {
	repo := memDevices{}
	uc := NewAppDevices(repo, &fakeRelay{}, activeSessions{}, clk{})
	in := AppDeviceInput{Name: "Pixel", Platform: domain.PushFCM, Token: "tok", PublicKey: make([]byte, 32)}
	d, err := uc.Register(appCtx("owner", "s1"), in)
	if err != nil || d.Session != "s1" || d.AccountID != "acc" || len(repo) != 1 {
		t.Fatal(d, err)
	}
	if _, err := uc.Register(appCtx("owner", ""), in); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatal("browser session:", err)
	}
	if _, err := uc.Register(appCtx("member", "s2"), in); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal("member:", err)
	}
	bad := in
	bad.PublicKey = []byte{1}
	if _, err := uc.Register(appCtx("admin", "s3"), bad); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("bad key:", err)
	}
	if err := uc.Unregister(appCtx("owner", "s1")); err != nil || len(repo) != 0 {
		t.Fatal(err, len(repo))
	}
	off := NewAppDevices(repo, nil, activeSessions{}, clk{})
	if _, err := off.Register(appCtx("owner", "s1"), in); !errors.Is(err, kernel.ErrUnavailable) {
		t.Fatal("push off:", err)
	}
}

func TestAppDevicesPublish(t *testing.T) {
	repo := memDevices{
		"live": {Session: "live"}, "revoked": {Session: "revoked"}, "gone": {Session: "gone"},
	}
	relay := &fakeRelay{gone: map[string]bool{"gone": true}}
	uc := NewAppDevices(repo, relay, activeSessions{"live": true, "gone": true}, clk{})
	if err := uc.Publish(context.Background(), domain.Channel{}, domain.Message{Title: "x"}); err != nil {
		t.Fatal(err)
	}
	if len(relay.pushed) != 1 || relay.pushed[0] != "live" {
		t.Fatal("pushed:", relay.pushed)
	}
	if _, ok := repo["revoked"]; ok {
		t.Fatal("device of a revoked session must be removed")
	}
	if _, ok := repo["gone"]; ok {
		t.Fatal("device the push service dropped must be removed")
	}
	if _, err := uc.List(appCtx("member", "")); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal(err)
	}
	if err := uc.Remove(appCtx("owner", ""), "live"); err != nil || len(repo) != 0 {
		t.Fatal(err)
	}
}

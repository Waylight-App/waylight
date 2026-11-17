package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/auth/domain"
)

// fakeDirectory knows "anna" with password "dir-secret" as member m1.
type fakeDirectory struct {
	calls   int
	down    bool
	account string
}

func (f *fakeDirectory) Authenticate(_ context.Context, name, password string) (string, error) {
	f.calls++
	if f.down {
		return "", errors.New("directory unreachable")
	}
	if name != "anna" {
		return "", ErrNotInDirectory
	}
	if password != "dir-secret" {
		return "", errors.New("invalid credentials")
	}
	return f.account, nil
}

func TestDirectoryLogin(t *testing.T) {
	e := newEnv().withAccount()
	anna, _ := domain.NewMember("m1", "anna", "hash:random-unknown", e.clock.now)
	guest, _ := domain.NewGuest("g1", "oma", "hash:"+goodPassword, nil, e.clock.now)
	e.accounts.guests = map[string]domain.Account{"m1": anna, "g1": guest}
	dir := &fakeDirectory{account: "m1"}
	e.deps.Directory = &DirectoryLogin{}
	e.deps.Directory.Set(dir)
	login := NewLogin(e.deps)
	try := func(name, password string) error {
		_, err := login.Execute(ctx, LoginRequest{Username: name, Password: password, ClientKey: "10.0.0.5"})
		return err
	}

	if err := try("anna", "dir-secret"); err != nil {
		t.Fatalf("directory member: %v", err)
	}
	if e.sessions.lastAccount() != "m1" {
		t.Fatalf("session for %q", e.sessions.lastAccount())
	}
	if err := try("anna", "wrong"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("wrong directory password: %v", err)
	}
	// Names the directory does not know are checked locally as before.
	if err := try("oma", goodPassword); err != nil {
		t.Fatalf("local guest: %v", err)
	}
	// The owner signs in locally and never reaches the directory, even when it is down.
	dir.down, dir.calls = true, 0
	if err := try("", goodPassword); err != nil || dir.calls != 0 {
		t.Fatalf("owner: %v, directory asked %d times", err, dir.calls)
	}
	if err := try("anna", "dir-secret"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("directory down: %v", err)
	}
	// The directory may only hand out members.
	dir.down, dir.account = false, "g1"
	if err := try("anna", "dir-secret"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("directory naming a guest: %v", err)
	}
}

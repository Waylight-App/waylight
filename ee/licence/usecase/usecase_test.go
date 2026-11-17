package usecase

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"waylight/ee/licence/domain"
	"waylight/internal/shared/kernel"
)

type memStore struct {
	key     string
	failing bool
}

func (m *memStore) Load(context.Context) (string, error) {
	if m.failing {
		return "", errors.New("disk on fire")
	}
	return m.key, nil
}
func (m *memStore) Save(_ context.Context, k string) error { m.key = k; return nil }
func (m *memStore) Delete(context.Context) error           { m.key = ""; return nil }

type counter struct{ n int }

func (c *counter) CountLicensedAccounts(context.Context) (int, error) { return c.n, nil }

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

var (
	admin  = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner"})
	member = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "m1", Role: "member"})
)

type fixture struct {
	svc      *Service
	store    *memStore
	accounts *counter
	priv     ed25519.PrivateKey
}

func setup(t *testing.T, env string) fixture {
	t.Helper()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	f := fixture{store: &memStore{}, accounts: &counter{n: 3}, priv: priv}
	f.svc = New(Deps{Store: f.store, Accounts: f.accounts, Clock: fixedClock{time.Date(2027, 1, 1, 12, 0, 0, 0, time.UTC)},
		PublicKey: pub, EnvKey: env})
	return f
}

func (f fixture) key(until string, seats int, features ...string) string {
	u, _ := time.Parse(time.DateOnly, until)
	return domain.Sign(domain.Licence{ID: "k", Licensee: "Example", Issued: u.AddDate(-1, 0, 0), Until: u, Seats: seats,
		Features: features}, f.priv)
}

func TestWithoutAKeyNothingIsLicensed(t *testing.T) {
	f := setup(t, "")
	if got := f.svc.Level(member, domain.FeatureBranding); got != domain.LevelNone {
		t.Errorf("level %d", got)
	}
	v, err := f.svc.Status(admin)
	if err != nil || v.Status != domain.StatusMissing {
		t.Errorf("status %+v, %v", v, err)
	}
}

func TestInstallUnlocksTheNamedFeaturesForEveryAccount(t *testing.T) {
	f := setup(t, "")
	v, err := f.svc.Install(admin, f.key("2027-12-31", 5, domain.FeatureBranding))
	if err != nil || v.Status != domain.StatusValid || v.Source != SourceSettings || v.Accounts != 3 {
		t.Fatalf("install: %+v, %v", v, err)
	}
	if got := f.svc.Level(member, domain.FeatureBranding); got != domain.LevelFull {
		t.Errorf("branding %d", got)
	}
	if got := f.svc.Level(member, domain.FeatureSSO); got != domain.LevelNone {
		t.Errorf("sso %d", got)
	}
}

func TestAnInvalidKeyIsRefusedAndTheOldOneStays(t *testing.T) {
	f := setup(t, "")
	good := f.key("2027-12-31", 5, domain.FeatureBranding)
	if _, err := f.svc.Install(admin, good); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Install(admin, good[:len(good)-4]+"AAAA"); !errors.Is(err, kernel.ErrInvalid) {
		t.Errorf("tampered key: %v", err)
	}
	if f.store.key != good {
		t.Error("the valid key was replaced")
	}
}

func TestMoreAccountsThanSeatsMakesFeaturesReadOnly(t *testing.T) {
	f := setup(t, "")
	if _, err := f.svc.Install(admin, f.key("2027-12-31", 3, domain.FeatureBranding)); err != nil {
		t.Fatal(err)
	}
	f.accounts.n = 4 // a fourth account joined: the key covers every account or changes stop
	if got := f.svc.Level(member, domain.FeatureBranding); got != domain.LevelRead {
		t.Errorf("level %d", got)
	}
	if v, _ := f.svc.Status(admin); v.Status != domain.StatusOverSeats {
		t.Errorf("status %s", v.Status)
	}
}

func TestAnExpiredKeyKeepsDataReadable(t *testing.T) {
	f := setup(t, "")
	if _, err := f.svc.Install(admin, f.key("2026-12-31", 5, domain.FeatureBranding)); err != nil {
		t.Fatal(err)
	}
	if got := f.svc.Level(member, domain.FeatureBranding); got != domain.LevelRead {
		t.Errorf("level %d", got)
	}
}

func TestTheEnvironmentKeyWinsAndCannotBeChangedHere(t *testing.T) {
	f := setup(t, "")
	env := f.key("2027-12-31", 5, domain.FeatureSSO)
	f = fixture{svc: New(Deps{Store: f.store, Accounts: f.accounts, Clock: fixedClock{time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		PublicKey: f.svc.d.PublicKey, EnvKey: env}), store: f.store, accounts: f.accounts, priv: f.priv}
	f.store.key = f.key("2027-12-31", 5, domain.FeatureBranding)
	if got := f.svc.Level(member, domain.FeatureSSO); got != domain.LevelFull {
		t.Errorf("env key not used: %d", got)
	}
	if _, err := f.svc.Install(admin, f.key("2027-12-31", 9, domain.FeatureSSO)); !errors.Is(err, ErrFromEnvironment) {
		t.Errorf("install: %v", err)
	}
	if err := f.svc.Remove(admin); !errors.Is(err, ErrFromEnvironment) {
		t.Errorf("remove: %v", err)
	}
	if v, _ := f.svc.Status(admin); v.Source != SourceEnvironment {
		t.Errorf("source %s", v.Source)
	}
}

func TestOnlyAdminsSeeOrChangeTheKey(t *testing.T) {
	f := setup(t, "")
	if _, err := f.svc.Status(member); !errors.Is(err, kernel.ErrForbidden) {
		t.Errorf("status: %v", err)
	}
	if _, err := f.svc.Install(member, f.key("2027-12-31", 5)); !errors.Is(err, kernel.ErrForbidden) {
		t.Errorf("install: %v", err)
	}
	if err := f.svc.Remove(member); !errors.Is(err, kernel.ErrForbidden) {
		t.Errorf("remove: %v", err)
	}
}

func TestRemoveHidesTheFeaturesAgain(t *testing.T) {
	f := setup(t, "")
	if _, err := f.svc.Install(admin, f.key("2027-12-31", 5, domain.FeatureBranding)); err != nil {
		t.Fatal(err)
	}
	if err := f.svc.Remove(admin); err != nil {
		t.Fatal(err)
	}
	if got := f.svc.Level(member, domain.FeatureBranding); got != domain.LevelNone {
		t.Errorf("level %d", got)
	}
}

func TestAStoreErrorLocksOnlyUntilTheStoreWorksAgain(t *testing.T) {
	f := setup(t, "")
	f.store.key = f.key("2027-12-31", 5, domain.FeatureBranding)
	f.store.failing = true
	if got := f.svc.Level(member, domain.FeatureBranding); got != domain.LevelNone {
		t.Errorf("level during failure %d", got)
	}
	f.store.failing = false
	if got := f.svc.Level(member, domain.FeatureBranding); got != domain.LevelFull {
		t.Errorf("level after recovery %d", got)
	}
}

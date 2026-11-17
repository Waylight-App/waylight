// Package usecase holds the application logic of licence keys (#307).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"
	"sync"

	"waylight/ee/licence/domain"
	"waylight/internal/shared/kernel"
)

// Store keeps the key entered in the settings (output port).
type Store interface {
	// Load returns the stored key, or "" when there is none.
	Load(ctx context.Context) (string, error)
	Save(ctx context.Context, key string) error
	Delete(ctx context.Context) error
}

// AccountCounter counts the accounts a key has to cover: the owner, members
// and guests that are not disabled, without kiosk displays (output port,
// implemented by the auth module).
type AccountCounter interface {
	CountLicensedAccounts(ctx context.Context) (int, error)
}

// Licensing is what ee/ features ask (input port): how far may a feature be
// used right now. It never fails; anything unexpected means LevelNone.
type Licensing interface {
	Level(ctx context.Context, feature string) domain.Level
}

// Errors of the settings use cases.
var (
	ErrFromEnvironment = kernel.NewPublicError(kernel.ErrConflict, "The licence key is set with the environment variable LICENCE_KEY and cannot be changed here.")
	errBadKey          = kernel.NewPublicError(kernel.ErrInvalid, "This licence key is not valid. Paste it exactly as you received it.")
	errAdminOnly       = kernel.NewPublicError(kernel.ErrForbidden, "Only an admin can see or change the licence key.")
)

func requireAdmin(ctx context.Context) error {
	if p := kernel.PrincipalFrom(ctx); !p.IsAdmin() {
		return errAdminOnly
	}
	return nil
}

// Source says where the key in use comes from.
type Source string

// The sources of a key.
const (
	SourceNone        Source = ""
	SourceEnvironment Source = "environment"
	SourceSettings    Source = "settings"
)

// View is the licence as admins see it.
type View struct {
	Status   domain.Status
	Source   Source
	Licence  *domain.Licence
	Accounts int
	Unknown  []string
	// Problem explains StatusInvalid, e.g. a key from another build.
	Problem string
}

// Deps are the ports of the use cases. PublicKey is compiled in; EnvKey comes
// from LICENCE_KEY and wins over a stored key.
type Deps struct {
	Store     Store
	Accounts  AccountCounter
	Clock     kernel.Clock
	PublicKey ed25519.PublicKey
	EnvKey    string
}

// Service implements Licensing and the admin use cases.
type Service struct {
	d Deps

	mu     sync.Mutex
	loaded bool
	lic    *domain.Licence
	source Source
	err    error
}

// New returns the service.
func New(d Deps) *Service { return &Service{d: d} }

// current returns the parsed key in use. The parse result is kept until the
// key changes, so a request pays for a signature check once, not every time.
func (s *Service) current(ctx context.Context) (*domain.Licence, Source, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loaded {
		return s.lic, s.source, s.err
	}
	key, source := strings.TrimSpace(s.d.EnvKey), SourceEnvironment
	if key == "" {
		stored, err := s.d.Store.Load(ctx)
		if err != nil {
			// Not cached: a database hiccup must not lock features until a restart.
			return nil, SourceNone, fmt.Errorf("load licence key: %w", err)
		}
		key, source = stored, SourceSettings
	}
	s.loaded, s.lic, s.source, s.err = true, nil, SourceNone, nil
	if key == "" {
		return nil, SourceNone, nil
	}
	s.source = source
	l, err := domain.Parse(key, s.d.PublicKey)
	if err != nil {
		s.err = err
		return nil, source, err
	}
	s.lic = &l
	return s.lic, source, nil
}

func (s *Service) forget() {
	s.mu.Lock()
	s.loaded = false
	s.mu.Unlock()
}

// Level implements Licensing.
func (s *Service) Level(ctx context.Context, feature string) domain.Level {
	l, _, err := s.current(ctx)
	if err != nil || l == nil {
		return domain.LevelNone
	}
	accounts, err := s.d.Accounts.CountLicensedAccounts(ctx)
	if err != nil {
		return domain.LevelNone
	}
	return domain.LevelOf(l, domain.Evaluate(l, s.d.Clock.Now(), accounts), feature)
}

// Levels returns the level of every known feature, for the interface.
func (s *Service) Levels(ctx context.Context) map[string]domain.Level {
	out := make(map[string]domain.Level, len(domain.KnownFeatures))
	for _, f := range domain.KnownFeatures {
		out[f] = s.Level(ctx, f)
	}
	return out
}

// Status returns the licence for the settings page.
func (s *Service) Status(ctx context.Context) (View, error) {
	if err := requireAdmin(ctx); err != nil {
		return View{}, err
	}
	accounts, err := s.d.Accounts.CountLicensedAccounts(ctx)
	if err != nil {
		return View{}, fmt.Errorf("count accounts: %w", err)
	}
	l, source, err := s.current(ctx)
	v := View{Source: source, Accounts: accounts}
	switch {
	case errors.Is(err, domain.ErrMalformed), errors.Is(err, domain.ErrSignature):
		v.Status, v.Problem = domain.StatusInvalid, err.Error()
	case err != nil:
		return View{}, err
	default:
		v.Status, v.Licence = domain.Evaluate(l, s.d.Clock.Now(), accounts), l
		if l != nil {
			v.Unknown = l.UnknownFeatures()
		}
	}
	return v, nil
}

// Install checks a key and stores it; an invalid key is refused and the old
// one stays.
func (s *Service) Install(ctx context.Context, key string) (View, error) {
	if err := requireAdmin(ctx); err != nil {
		return View{}, err
	}
	if strings.TrimSpace(s.d.EnvKey) != "" {
		return View{}, ErrFromEnvironment
	}
	key = strings.TrimSpace(key)
	if _, err := domain.Parse(key, s.d.PublicKey); err != nil {
		return View{}, errBadKey
	}
	if err := s.d.Store.Save(ctx, key); err != nil {
		return View{}, fmt.Errorf("save licence key: %w", err)
	}
	s.forget()
	return s.Status(ctx)
}

// Remove deletes the stored key; the licensed features are hidden again,
// their data stays.
func (s *Service) Remove(ctx context.Context) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(s.d.EnvKey) != "" {
		return ErrFromEnvironment
	}
	if err := s.d.Store.Delete(ctx); err != nil {
		return fmt.Errorf("delete licence key: %w", err)
	}
	s.forget()
	return nil
}

package usecase

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"

	"waylight/internal/modules/integrations/domain"
	"waylight/internal/shared/kernel"
)

// Saved is a connection a module keeps in its own settings (see
// modules.SavedConnection).
type Saved struct {
	Source, Service, Name, URL string
	Credential                 map[string]string
}

// CopyRepository records which module settings were copied (output port).
type CopyRepository interface {
	Repository
	// Copied returns the sources copied so far.
	Copied(ctx context.Context) (map[string]bool, error)
	// SaveCopy writes c and records source atomically, after verify
	// accepted the connection as read back from the store. A source that
	// is recorded already is left alone.
	SaveCopy(ctx context.Context, source string, c domain.Connection, verify func(domain.Connection) error) error
}

// CopyResult says what a run did: sources copied now, and sources skipped
// with the reason. Reasons never contain credential values.
type CopyResult struct {
	Copied  []string
	Skipped map[string]string
}

// Copier copies the module settings into connections (#610, step 2 of
// the plan). The module settings are only read; widgets keep using them
// until they read connections.
type Copier interface {
	Copy(ctx context.Context, saved []Saved) (CopyResult, error)
}

// ErrCopyMismatch means a copy read back from the store differs from its
// source; the copy is rolled back and tried again on the next run.
var ErrCopyMismatch = errors.New("the copy differs from the module settings")

type copier struct {
	repo  CopyRepository
	specs domain.Specs
	clock kernel.Clock
	ids   kernel.IDGenerator
}

// NewCopier returns the Copier for the given services.
func NewCopier(repo CopyRepository, services []Service, clock kernel.Clock, ids kernel.IDGenerator) Copier {
	list := make([]domain.Spec, 0, len(services))
	for _, s := range services {
		list = append(list, s.Spec)
	}
	return &copier{repo: repo, specs: domain.NewSpecs(list), clock: clock, ids: ids}
}

// Copy implements Copier. Every source is copied at most once; a source
// that cannot be copied is skipped and tried again on the next run.
func (c *copier) Copy(ctx context.Context, saved []Saved) (CopyResult, error) {
	res := CopyResult{Skipped: map[string]string{}}
	done, err := c.repo.Copied(ctx)
	if err != nil {
		return res, fmt.Errorf("list copies: %w", err)
	}
	for _, s := range saved {
		if done[s.Source] {
			continue
		}
		if err := c.copyOne(ctx, s); err != nil {
			res.Skipped[s.Source] = reason(err)
			continue
		}
		done[s.Source] = true
		res.Copied = append(res.Copied, s.Source)
	}
	return res, nil
}

func (c *copier) copyOne(ctx context.Context, s Saved) error {
	spec, ok := c.specs[s.Service]
	if !ok {
		return domain.ErrUnknownService
	}
	cred, err := domain.NewCredential(s.Credential)
	if err != nil {
		return err
	}
	if err := spec.CheckCredential(cred, true); err != nil {
		return err
	}
	list, err := c.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("list connections: %w", err)
	}
	if len(list) >= domain.MaxConnections {
		return domain.ErrTooMany
	}
	name := s.Name
	if name == "" {
		name = spec.Title
	}
	conn, err := domain.NewConnection(c.ids.NewID(), s.Service, name, fixedURL(spec, s.URL), cred, c.clock.Now())
	if err != nil {
		return err
	}
	// The store normalises addresses; one that would point elsewhere after
	// that is not copied, so the credential never reaches another host.
	if spec.FixedURL == "" && kernel.TargetChanged([]string{s.URL}, []string{conn.URL}) {
		return ErrCopyMismatch
	}
	conn.Name = freeName(list, conn)
	want := conn
	return c.repo.SaveCopy(ctx, s.Source, conn, func(got domain.Connection) error {
		if got.Service != want.Service || got.URL != want.URL || got.Name != want.Name || !maps.Equal(got.Credential, want.Credential) {
			return ErrCopyMismatch
		}
		return nil
	})
}

// freeName keeps the name unless another connection of the service has it
// already, then numbers it: "Immich 2", "Immich 3"…
func freeName(list []domain.Connection, c domain.Connection) string {
	base := c.Name
	for n := 2; taken(list, c); n++ {
		suffix := " " + strconv.Itoa(n)
		c.Name = truncate(base, domain.MaxNameLength-len(suffix)) + suffix
	}
	return c.Name
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max])
	}
	return s
}

// reason turns an error into a short text for the log. Validation errors
// name the field, other errors their wrapping text; none of the errors
// along the copy carries a credential value.
func reason(err error) string {
	var v *kernel.ValidationError
	if errors.As(err, &v) {
		return v.Field + ": " + v.Message
	}
	return err.Error()
}

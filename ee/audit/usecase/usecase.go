// Package usecase records changes and lets admins read, export and age out
// the audit log (#303).
//
// Ring: Use cases.
package usecase

import (
	"context"
	"fmt"
	"time"

	"waylight/ee/audit/domain"
	licence "waylight/ee/licence/domain"
	"waylight/internal/shared/kernel"
)

// Store keeps entries (output port). There is no update and no delete of a
// single entry: the log can only grow and age out.
type Store interface {
	Append(ctx context.Context, e domain.Entry) error
	List(ctx context.Context, f domain.Filter) ([]domain.Entry, error)
	DeleteBefore(ctx context.Context, t time.Time) (int64, error)
	Retention(ctx context.Context) (int, error)
	SetRetention(ctx context.Context, days int) error
}

// Licensing tells whether the audit log is licensed (output port).
type Licensing interface {
	Level(ctx context.Context, feature string) licence.Level
}

var errAdminOnly = kernel.NewPublicError(kernel.ErrForbidden, "Only an admin can read the audit log.")

func requireAdmin(ctx context.Context) error {
	if p := kernel.PrincipalFrom(ctx); !p.IsAdmin() {
		return errAdminOnly
	}
	return nil
}

// Deps are the ports.
type Deps struct {
	Store     Store
	Clock     kernel.Clock
	IDs       kernel.IDGenerator
	Licensing Licensing
}

// Service implements the use cases.
type Service struct{ d Deps }

// New returns the service.
func New(d Deps) *Service { return &Service{d: d} }

// Record appends a change. It records while the key is valid or expired (the
// log is evidence, and keeping it costs nothing), but not without a key.
// Recording never fails the request that made the change.
func (s *Service) Record(ctx context.Context, e domain.Entry) error {
	if s.d.Licensing.Level(ctx, licence.FeatureAuditLog) == licence.LevelNone {
		return nil
	}
	e.ID, e.At = s.d.IDs.NewID(), s.d.Clock.Now().UTC()
	if err := s.d.Store.Append(ctx, e); err != nil {
		return fmt.Errorf("append audit entry: %w", err)
	}
	return nil
}

// List returns entries, newest first.
func (s *Service) List(ctx context.Context, f domain.Filter) ([]domain.Entry, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	if f.Limit <= 0 || f.Limit > domain.MaxLimit {
		f.Limit = domain.MaxLimit
	}
	return s.d.Store.List(ctx, f)
}

// Export returns all entries matching the filter, without a page limit.
func (s *Service) Export(ctx context.Context, f domain.Filter) ([]domain.Entry, error) {
	if err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	f.Limit = 0
	return s.d.Store.List(ctx, f)
}

// Retention returns the retention in days.
func (s *Service) Retention(ctx context.Context) (int, error) {
	if err := requireAdmin(ctx); err != nil {
		return 0, err
	}
	return s.d.Store.Retention(ctx)
}

// SetRetention changes the retention; the change is itself recorded.
func (s *Service) SetRetention(ctx context.Context, days int) error {
	if err := requireAdmin(ctx); err != nil {
		return err
	}
	if err := domain.ValidRetention(days); err != nil {
		return err
	}
	return s.d.Store.SetRetention(ctx, days)
}

// Prune deletes entries older than the retention (background job).
func (s *Service) Prune(ctx context.Context) error {
	days, err := s.d.Store.Retention(ctx)
	if err != nil || days == 0 {
		return err
	}
	_, err = s.d.Store.DeleteBefore(ctx, s.d.Clock.Now().UTC().AddDate(0, 0, -days))
	return err
}

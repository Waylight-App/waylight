package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"waylight/internal/modules/backup/domain"
	"waylight/internal/shared/kernel"
)

// FileView is a backup at the use case boundary.
type FileView struct {
	Name      string
	Size      int64
	CreatedAt time.Time
}

// Backups is the input port of the module; every method is one use case.
type Backups interface {
	// RunScheduled takes a backup if one is due and prunes old ones (job).
	RunScheduled(ctx context.Context) error
	// BackupNow takes a backup immediately and prunes old ones.
	BackupNow(ctx context.Context) (FileView, error)
	List(ctx context.Context) ([]FileView, error)
	// Open returns the content of a backup for download.
	Open(ctx context.Context, name string) (io.ReadCloser, FileView, error)
	// Status returns the traffic light of the backups (#677).
	Status(ctx context.Context) (StatusView, error)
	SafetyCopier
}

// Snapshotter writes a consistent copy of the live database (output port).
type Snapshotter interface {
	SnapshotTo(ctx context.Context, path string) error
}

// Store manages backup files (output port).
type Store interface {
	// Path returns the absolute path for a new backup file name.
	Path(name string) string
	List(ctx context.Context) ([]domain.File, error)
	Open(name string) (io.ReadCloser, domain.File, error)
	Remove(name string) error
}

type backups struct {
	policy  domain.Policy
	snap    Snapshotter
	store   Store
	clock   kernel.Clock
	after   func(ctx context.Context, f FileView)
	tracker *Tracker
}

// Option configures the backup use cases.
type Option func(*backups)

// WithAfterBackup registers a hook called after each successful backup,
// e.g. the off-site upload (#46).
func WithAfterBackup(fn func(ctx context.Context, f FileView)) Option {
	return func(b *backups) { b.after = fn }
}

// WithTracker shares the outcome tracker with the off-site upload, which
// records its results there too.
func WithTracker(t *Tracker) Option {
	return func(b *backups) { b.tracker = t }
}

// NewBackups returns the backup use cases.
func NewBackups(policy domain.Policy, snap Snapshotter, store Store, clock kernel.Clock, opts ...Option) Backups {
	b := &backups{policy: policy, snap: snap, store: store, clock: clock, tracker: &Tracker{}}
	for _, o := range opts {
		o(b)
	}
	return b
}

func toView(f domain.File) FileView {
	return FileView{Name: f.Name, Size: f.Size, CreatedAt: f.CreatedAt}
}

// RunScheduled implements Backups.
func (uc *backups) RunScheduled(ctx context.Context) error {
	files, err := uc.store.List(ctx)
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	var latest time.Time
	for _, f := range files {
		if f.CreatedAt.After(latest) {
			latest = f.CreatedAt
		}
	}
	if !uc.policy.Due(latest, uc.clock.Now()) {
		return nil
	}
	_, err = uc.BackupNow(ctx)
	return err
}

// BackupNow implements Backups. Its outcome feeds the traffic light.
func (uc *backups) BackupNow(ctx context.Context) (FileView, error) {
	now := uc.clock.Now()
	f, err := uc.backup(ctx, now)
	uc.tracker.BackupResult(now, err)
	if err != nil {
		return FileView{}, err
	}
	if uc.after != nil {
		uc.after(ctx, f)
	}
	return f, nil
}

// SafetyCopy implements SafetyCopier: a backup of the current state before
// a restore. It prunes nothing and skips the off-site upload, which the
// restart right after it would cut off.
func (uc *backups) SafetyCopy(ctx context.Context) (FileView, error) {
	return uc.snapshot(ctx, uc.clock.Now())
}

func (uc *backups) backup(ctx context.Context, now time.Time) (FileView, error) {
	f, err := uc.snapshot(ctx, now)
	if err != nil {
		return FileView{}, err
	}
	return f, uc.prune(ctx)
}

func (uc *backups) snapshot(ctx context.Context, now time.Time) (FileView, error) {
	name, err := uc.freeName(now)
	if err != nil {
		return FileView{}, err
	}
	if err := uc.snap.SnapshotTo(ctx, uc.store.Path(name)); err != nil {
		return FileView{}, fmt.Errorf("create backup: %w", err)
	}
	rc, f, err := uc.store.Open(name)
	if err != nil {
		return FileView{}, err
	}
	_ = rc.Close()
	return toView(f), nil
}

// maxNameShift bounds the search for a free file name.
const maxNameShift = 60

// freeName returns the file name for a backup taken at now. Names have
// second precision, so a backup right after another one (a restore's safety
// copy after "Back up now") moves to the next free second instead of
// failing on the existing file.
func (uc *backups) freeName(now time.Time) (string, error) {
	for i := range maxNameShift {
		name := domain.FileName(now.Add(time.Duration(i) * time.Second))
		rc, _, err := uc.store.Open(name)
		if errors.Is(err, domain.ErrBackupNotFound) {
			return name, nil
		}
		if err != nil {
			return "", fmt.Errorf("check backup name: %w", err)
		}
		_ = rc.Close()
	}
	return "", errors.New("no free backup file name")
}

func (uc *backups) prune(ctx context.Context) error {
	files, err := uc.store.List(ctx)
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	for _, f := range uc.policy.ToPrune(files) {
		if err := uc.store.Remove(f.Name); err != nil {
			return fmt.Errorf("remove old backup: %w", err)
		}
	}
	return nil
}

// List implements Backups; newest first.
func (uc *backups) List(ctx context.Context) ([]FileView, error) {
	files, err := uc.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	out := make([]FileView, 0, len(files))
	for i := len(files) - 1; i >= 0; i-- {
		out = append(out, toView(files[i]))
	}
	return out, nil
}

// Open implements Backups. Names are validated so requests cannot escape
// the backup directory.
func (uc *backups) Open(_ context.Context, name string) (io.ReadCloser, FileView, error) {
	if !domain.ValidName(name) {
		return nil, FileView{}, domain.ErrBackupNotFound
	}
	rc, f, err := uc.store.Open(name)
	if err != nil {
		return nil, FileView{}, err
	}
	return rc, toView(f), nil
}

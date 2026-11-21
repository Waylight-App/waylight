package usecase

import (
	"context"
	"fmt"
	"sync"
	"time"

	"waylight/internal/modules/focus/domain"
	"waylight/internal/shared/kernel"
)

// Focus is the input port of the module; every method is one use case
// (#58).
type Focus interface {
	// RecordSession stores a finished focus phase.
	RecordSession(ctx context.Context, minutes int, endedAt time.Time) (domain.Session, error)
	// WeeklyStats counts sessions per day for the last seven days.
	WeeklyStats(ctx context.Context) ([]domain.Day, error)
	// Prune deletes sessions older than the retention (job).
	Prune(ctx context.Context) error
	// Timer returns the account's timer; a phase that ran out is advanced
	// and, for a focus phase, recorded as a session (#631).
	Timer(ctx context.Context) (domain.Timer, error)
	// ControlTimer starts, pauses, resets or skips the account's timer.
	ControlTimer(ctx context.Context, action TimerAction) (domain.Timer, error)
	// ConfigureTimer stores the phase lengths of the widget's settings.
	ConfigureTimer(ctx context.Context, c domain.TimerConfig) (domain.Timer, error)
}

// Repository persists sessions (output port, implemented in
// adapter/gateway).
type Repository interface {
	Save(ctx context.Context, s domain.Session) error
	Since(ctx context.Context, t time.Time) ([]domain.Session, error)
	DeleteBefore(ctx context.Context, t time.Time) error
}

// TimerRepository keeps one timer per account (output port, implemented in
// adapter/gateway). Timer reports false when the account has none yet.
type TimerRepository interface {
	Timer(ctx context.Context) (domain.Timer, bool, error)
	SaveTimer(ctx context.Context, t domain.Timer) error
}

// Retention is how long sessions are kept.
const Retention = 90 * 24 * time.Hour

// Deps bundles the output ports. Location is the user's time zone
// (nil = time.Local).
type Deps struct {
	Repo     Repository
	Timers   TimerRepository
	Clock    kernel.Clock
	IDs      kernel.IDGenerator
	Location *time.Location
}

type uc struct {
	d Deps
	// timerMu serialises reading, advancing and saving timers, so two
	// devices that look at a finished phase at once record it only once.
	timerMu sync.Mutex
}

// NewFocus returns the focus timer use cases.
func NewFocus(d Deps) Focus {
	if d.Location == nil {
		d.Location = time.Local
	}
	return &uc{d: d}
}

// RecordSession implements Focus.
func (u *uc) RecordSession(ctx context.Context, minutes int, endedAt time.Time) (domain.Session, error) {
	s, err := domain.NewSession(u.d.IDs.NewID(), minutes, endedAt, u.d.Clock.Now())
	if err != nil {
		return domain.Session{}, err
	}
	if err := u.d.Repo.Save(ctx, s); err != nil {
		return domain.Session{}, fmt.Errorf("save session: %w", err)
	}
	return s, nil
}

// WeeklyStats implements Focus.
func (u *uc) WeeklyStats(ctx context.Context) ([]domain.Day, error) {
	now := u.d.Clock.Now().In(u.d.Location)
	sessions, err := u.d.Repo.Since(ctx, domain.WeekStart(now))
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return domain.Week(sessions, now), nil
}

// Prune implements Focus.
func (u *uc) Prune(ctx context.Context) error {
	return u.d.Repo.DeleteBefore(ctx, u.d.Clock.Now().Add(-Retention))
}

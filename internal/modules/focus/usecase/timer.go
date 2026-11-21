package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"waylight/internal/modules/focus/domain"
	"waylight/internal/shared/kernel"
)

// TimerAction is a control of the timer.
type TimerAction string

// The controls of the widget; the compact view offers start and pause.
const (
	TimerStart TimerAction = "start"
	TimerPause TimerAction = "pause"
	TimerReset TimerAction = "reset"
	TimerSkip  TimerAction = "skip"
)

// errNoTimer is returned to guests, kiosk displays and the system, which
// have no timer of their own.
var errNoTimer = kernel.NewPublicError(kernel.ErrForbidden, "Only a signed-in account has its own focus timer.")

func hasOwnTimer(p kernel.Principal) bool { return !p.TabScoped() && !p.IsSystem() }

// Timer implements Focus. Viewers without a timer of their own see an
// untouched one, so the widget and its compact view still render.
func (u *uc) Timer(ctx context.Context) (domain.Timer, error) {
	if !hasOwnTimer(kernel.PrincipalFrom(ctx)) {
		return domain.NewTimer(domain.DefaultTimerConfig), nil
	}
	return u.changeTimer(ctx, nil)
}

// ControlTimer implements Focus.
func (u *uc) ControlTimer(ctx context.Context, action TimerAction) (domain.Timer, error) {
	var change func(domain.Timer, time.Time) domain.Timer
	switch action {
	case TimerStart:
		change = domain.Timer.Start
	case TimerPause:
		change = domain.Timer.Pause
	case TimerReset:
		change = func(t domain.Timer, _ time.Time) domain.Timer { return t.Reset() }
	case TimerSkip:
		change = func(t domain.Timer, _ time.Time) domain.Timer { return t.Skip() }
	default:
		return domain.Timer{}, kernel.NewValidationError("action", "Unknown timer action.")
	}
	if !hasOwnTimer(kernel.PrincipalFrom(ctx)) {
		return domain.Timer{}, errNoTimer
	}
	return u.changeTimer(ctx, change)
}

// ConfigureTimer implements Focus.
func (u *uc) ConfigureTimer(ctx context.Context, c domain.TimerConfig) (domain.Timer, error) {
	if !hasOwnTimer(kernel.PrincipalFrom(ctx)) {
		return domain.Timer{}, errNoTimer
	}
	return u.changeTimer(ctx, func(t domain.Timer, _ time.Time) domain.Timer { return t.WithConfig(c) })
}

// changeTimer loads the timer, advances a phase that ran out (recording a
// finished focus phase), applies change and saves the result if anything
// changed.
func (u *uc) changeTimer(ctx context.Context, change func(domain.Timer, time.Time) domain.Timer) (domain.Timer, error) {
	u.timerMu.Lock()
	defer u.timerMu.Unlock()
	stored, ok, err := u.d.Timers.Timer(ctx)
	if err != nil {
		return domain.Timer{}, fmt.Errorf("load timer: %w", err)
	}
	if !ok {
		stored = domain.NewTimer(domain.DefaultTimerConfig)
	}
	now := u.d.Clock.Now()
	t, finished := stored.Advance(now)
	if finished != nil && finished.Phase == domain.PhaseWork {
		if err := u.recordFinished(ctx, *finished, now); err != nil {
			return domain.Timer{}, err
		}
	}
	if change != nil {
		t = change(t, now)
	}
	// An account without a timer gets one stored only once it changes.
	if t.Equal(stored) {
		return t, nil
	}
	if err := u.d.Timers.SaveTimer(ctx, t); err != nil {
		return domain.Timer{}, fmt.Errorf("save timer: %w", err)
	}
	return t, nil
}

// recordFinished stores a focus phase that ran out. One that ended more
// than a day ago is dropped like a late report from a sleeping tab.
func (u *uc) recordFinished(ctx context.Context, f domain.Finished, now time.Time) error {
	s, err := domain.NewSession(u.d.IDs.NewID(), int(f.Length/time.Minute), f.At, now)
	if errors.Is(err, kernel.ErrInvalid) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := u.d.Repo.Save(ctx, s); err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

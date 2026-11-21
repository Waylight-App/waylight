package domain

import (
	"fmt"
	"math"
	"time"

	"waylight/internal/shared/kernel"
)

// Phase is a phase of the focus timer.
type Phase string

// The phases of a cycle: work, short breaks, and a long break after every
// LongEvery-th focus session.
const (
	PhaseWork  Phase = "work"
	PhaseShort Phase = "short"
	PhaseLong  Phase = "long"
)

// MaxLongEvery is the largest number of focus sessions before a long break.
const MaxLongEvery = 12

// TimerConfig holds the phase lengths the timer uses; the widget sends its
// settings, the compact view uses what was stored last.
type TimerConfig struct {
	Work, Short, Long time.Duration
	// LongEvery is the number of focus sessions before a long break.
	LongEvery int
}

// DefaultTimerConfig mirrors the widget's defaults.
var DefaultTimerConfig = TimerConfig{Work: 25 * time.Minute, Short: 5 * time.Minute, Long: 15 * time.Minute, LongEvery: 4}

// NewTimerConfig validates phase lengths in minutes.
func NewTimerConfig(work, short, long, longEvery int) (TimerConfig, error) {
	for field, m := range map[string]int{"work": work, "short": short, "long": long} {
		if m < 1 || m > MaxMinutes {
			return TimerConfig{}, kernel.NewValidationError(field, "A phase lasts between 1 and 180 minutes.")
		}
	}
	if longEvery < 1 || longEvery > MaxLongEvery {
		return TimerConfig{}, kernel.NewValidationError("longEvery", "A long break follows every 1 to 12 sessions.")
	}
	return TimerConfig{Work: minutes(work), Short: minutes(short), Long: minutes(long), LongEvery: longEvery}, nil
}

func minutes(n int) time.Duration { return time.Duration(n) * time.Minute }

func (c TimerConfig) duration(p Phase) time.Duration {
	switch p {
	case PhaseShort:
		return c.Short
	case PhaseLong:
		return c.Long
	default:
		return c.Work
	}
}

// Timer is the running or paused focus timer of an account (#631). It is
// kept on the server so the full widget, its compact view and other
// devices show and control the same timer. Like the browser timer it
// replaces, it is based on the end time, not on ticks.
type Timer struct {
	Phase Phase
	// Done counts the focus sessions finished in the current cycle.
	Done int
	// EndsAt is the end of the running phase; zero while paused.
	EndsAt time.Time
	// Left is the time left while paused.
	Left time.Duration
	// Total is the full length of the phase when it was prepared.
	Total  time.Duration
	Config TimerConfig
}

// NewTimer is a paused, untouched focus phase.
func NewTimer(c TimerConfig) Timer { return fresh(PhaseWork, 0, c) }

func fresh(p Phase, done int, c TimerConfig) Timer {
	d := c.duration(p)
	return Timer{Phase: p, Done: done, Left: d, Total: d, Config: c}
}

// Equal reports whether two timers are in the same state.
func (t Timer) Equal(o Timer) bool {
	return t.Phase == o.Phase && t.Done == o.Done && t.EndsAt.Equal(o.EndsAt) && t.Left == o.Left &&
		t.Total == o.Total && t.Config == o.Config
}

// Running reports whether the phase is counting down.
func (t Timer) Running() bool { return !t.EndsAt.IsZero() }

// Untouched reports whether the phase is paused and was never started.
func (t Timer) Untouched() bool { return !t.Running() && t.Left == t.Total }

// Remaining is the time left at now, never negative.
func (t Timer) Remaining(now time.Time) time.Duration {
	if !t.Running() {
		return t.Left
	}
	return max(0, t.EndsAt.Sub(now))
}

// Start lets a paused phase run; a running one is left alone.
func (t Timer) Start(now time.Time) Timer {
	if !t.Running() {
		t.EndsAt = now.Add(t.Left)
	}
	return t
}

// Pause stops a running phase and keeps the time left.
func (t Timer) Pause(now time.Time) Timer {
	if t.Running() {
		t.Left = t.Remaining(now)
		t.EndsAt = time.Time{}
	}
	return t
}

// Reset starts the current phase over, paused.
func (t Timer) Reset() Timer { return fresh(t.Phase, t.Done, t.Config) }

// Skip jumps to the next phase without counting a session.
func (t Timer) Skip() Timer {
	if t.Phase == PhaseWork {
		return fresh(PhaseShort, t.Done, t.Config)
	}
	p, done := t.next()
	return fresh(p, done, t.Config)
}

// WithConfig follows changed settings: a paused phase that was not started
// yet takes the new length; a started one keeps its time, capped at the new
// length; a running one is left as it is.
func (t Timer) WithConfig(c TimerConfig) Timer {
	t.Config = c
	if t.Running() {
		return t
	}
	d := c.duration(t.Phase)
	switch {
	case t.Left == t.Total:
		return fresh(t.Phase, t.Done, c)
	case t.Left > d:
		t.Left, t.Total = d, d
	}
	return t
}

func (t Timer) next() (Phase, int) {
	if t.Phase != PhaseWork {
		if t.Phase == PhaseLong {
			return PhaseWork, 0
		}
		return PhaseWork, t.Done
	}
	done := t.Done + 1
	if done%t.Config.LongEvery == 0 {
		return PhaseLong, done
	}
	return PhaseShort, done
}

// Finished is a phase that ran out.
type Finished struct {
	Phase Phase
	// At is when it ended, possibly long before anyone looked.
	At time.Time
	// Length is the full length of the phase.
	Length time.Duration
}

// Advance prepares the next phase, paused, once the running one has ended.
// It moves at most one phase, so a timer nobody looked at overnight does
// not run through several cycles.
func (t Timer) Advance(now time.Time) (Timer, *Finished) {
	if !t.Running() || now.Before(t.EndsAt) {
		return t, nil
	}
	p, done := t.next()
	return fresh(p, done, t.Config), &Finished{Phase: t.Phase, At: t.EndsAt, Length: t.Total}
}

// FormatClock writes a duration as m:ss, rounded up so 0:01 shows until the
// end, like the widget.
func FormatClock(d time.Duration) string {
	s := int(math.Ceil(d.Seconds()))
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

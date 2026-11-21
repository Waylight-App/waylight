package domain

import (
	"errors"
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

var t0 = time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)

func cfg(t *testing.T, work, short, long, every int) TimerConfig {
	t.Helper()
	c, err := NewTimerConfig(work, short, long, every)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestTimerConfigValidation(t *testing.T) {
	for _, c := range [][4]int{{0, 5, 15, 4}, {25, 181, 15, 4}, {25, 5, 15, 0}, {25, 5, 15, 13}} {
		if _, err := NewTimerConfig(c[0], c[1], c[2], c[3]); !errors.Is(err, kernel.ErrInvalid) {
			t.Errorf("%v accepted: %v", c, err)
		}
	}
}

func TestTimerStartPauseAndRemaining(t *testing.T) {
	tm := NewTimer(cfg(t, 25, 5, 15, 4))
	if tm.Running() || !tm.Untouched() || tm.Remaining(t0) != 25*time.Minute {
		t.Fatalf("new timer %+v", tm)
	}
	tm = tm.Start(t0)
	if again := tm.Start(t0.Add(time.Minute)); !again.Equal(tm) {
		t.Fatal("starting a running timer moved its end")
	}
	tm = tm.Pause(t0.Add(10 * time.Minute))
	if tm.Running() || tm.Untouched() || tm.Left != 15*time.Minute {
		t.Fatalf("paused %+v", tm)
	}
	if got := FormatClock(tm.Remaining(t0)); got != "15:00" {
		t.Errorf("clock %q", got)
	}
	if got := FormatClock(1500 * time.Millisecond); got != "0:02" {
		t.Errorf("clock rounds up: %q", got)
	}
}

func TestTimerAdvancesOnePhase(t *testing.T) {
	tm := NewTimer(cfg(t, 25, 5, 15, 2)).Start(t0)
	if same, f := tm.Advance(t0.Add(time.Minute)); f != nil || !same.Equal(tm) {
		t.Fatal("advanced before the end")
	}
	// Nobody looked for hours: only the focus phase finishes, the break waits.
	next, f := tm.Advance(t0.Add(5 * time.Hour))
	if f == nil || f.Phase != PhaseWork || !f.At.Equal(t0.Add(25*time.Minute)) || f.Length != 25*time.Minute {
		t.Fatalf("finished %+v", f)
	}
	if next.Phase != PhaseShort || next.Done != 1 || next.Running() || next.Left != 5*time.Minute {
		t.Fatalf("next %+v", next)
	}
	// The second session of the cycle leads to the long break, which resets the count.
	next, _ = next.Skip().Start(t0).Advance(t0.Add(time.Hour))
	if next.Phase != PhaseLong || next.Done != 2 {
		t.Fatalf("after second session %+v", next)
	}
	if after := next.Skip(); after.Phase != PhaseWork || after.Done != 0 {
		t.Fatalf("after long break %+v", after)
	}
}

func TestTimerSkipAndReset(t *testing.T) {
	tm := NewTimer(cfg(t, 25, 5, 15, 4)).Start(t0).Pause(t0.Add(time.Minute))
	if r := tm.Reset(); !r.Untouched() || r.Phase != PhaseWork {
		t.Fatalf("reset %+v", r)
	}
	if s := tm.Skip(); s.Phase != PhaseShort || s.Done != 0 {
		t.Fatalf("skipping focus counts no session: %+v", s)
	}
}

func TestTimerWithConfig(t *testing.T) {
	short := cfg(t, 10, 5, 15, 4)
	if tm := NewTimer(DefaultTimerConfig).WithConfig(short); tm.Left != 10*time.Minute || tm.Total != 10*time.Minute {
		t.Fatalf("untouched phase keeps old length: %+v", tm)
	}
	started := NewTimer(DefaultTimerConfig).Start(t0).Pause(t0.Add(5 * time.Minute))
	if tm := started.WithConfig(short); tm.Left != 10*time.Minute {
		t.Fatalf("started phase not capped: %+v", tm)
	}
	running := NewTimer(DefaultTimerConfig).Start(t0)
	if tm := running.WithConfig(short); !tm.EndsAt.Equal(running.EndsAt) || tm.Config != short {
		t.Fatalf("running phase changed: %+v", tm)
	}
}

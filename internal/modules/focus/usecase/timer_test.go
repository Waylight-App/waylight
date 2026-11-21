package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/focus/domain"
	"waylight/internal/shared/kernel"
)

type memTimers struct {
	t     *domain.Timer
	saves int
}

func (m *memTimers) Timer(context.Context) (domain.Timer, bool, error) {
	if m.t == nil {
		return domain.Timer{}, false, nil
	}
	return *m.t, true, nil
}

func (m *memTimers) SaveTimer(_ context.Context, t domain.Timer) error {
	m.t = &t
	m.saves++
	return nil
}

type movingClock struct{ t *time.Time }

func (c movingClock) Now() time.Time { return *c.t }

func TestTimerControlsAndRecordsFinishedFocus(t *testing.T) {
	ctx := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	repo, timers := &memRepo{}, &memTimers{}
	u := NewFocus(Deps{Repo: repo, Timers: timers, Clock: movingClock{&now}, IDs: ids{}, Location: time.UTC})

	// Reading an account without a timer does not store one.
	if tm, err := u.Timer(ctx); err != nil || !tm.Untouched() || timers.saves != 0 {
		t.Fatalf("first read %+v %v (%d saves)", tm, err, timers.saves)
	}
	if _, err := u.ControlTimer(ctx, "explode"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("unknown action: %v", err)
	}
	c, _ := domain.NewTimerConfig(1, 1, 1, 4)
	if _, err := u.ConfigureTimer(ctx, c); err != nil {
		t.Fatal(err)
	}
	tm, err := u.ControlTimer(ctx, TimerStart)
	if err != nil || !tm.Running() {
		t.Fatalf("start %+v %v", tm, err)
	}
	// Two reads after the end record the session once.
	now = now.Add(2 * time.Minute)
	for range 2 {
		if tm, err = u.Timer(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if tm.Phase != domain.PhaseShort || tm.Running() || len(repo.sessions) != 1 || repo.sessions[0].Minutes != 1 {
		t.Fatalf("after the end %+v, sessions %v", tm, repo.sessions)
	}
	// A finished break is not recorded.
	_, _ = u.ControlTimer(ctx, TimerStart)
	now = now.Add(2 * time.Minute)
	if tm, _ = u.Timer(ctx); tm.Phase != domain.PhaseWork || len(repo.sessions) != 1 {
		t.Fatalf("after the break %+v, sessions %v", tm, repo.sessions)
	}
	if tm, _ = u.ControlTimer(ctx, TimerSkip); tm.Phase != domain.PhaseShort {
		t.Fatalf("skip %+v", tm)
	}
	if tm, _ = u.ControlTimer(ctx, TimerReset); !tm.Untouched() {
		t.Fatalf("reset %+v", tm)
	}
}

func TestTimerDropsSessionsOlderThanADay(t *testing.T) {
	ctx := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	repo, timers := &memRepo{}, &memTimers{}
	u := NewFocus(Deps{Repo: repo, Timers: timers, Clock: movingClock{&now}, IDs: ids{}, Location: time.UTC})
	if _, err := u.ControlTimer(ctx, TimerStart); err != nil {
		t.Fatal(err)
	}
	now = now.Add(48 * time.Hour)
	if tm, err := u.Timer(ctx); err != nil || tm.Phase != domain.PhaseShort || len(repo.sessions) != 0 {
		t.Fatalf("%+v %v %v", tm, err, repo.sessions)
	}
}

func TestGuestsHaveNoTimer(t *testing.T) {
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	timers := &memTimers{}
	u := NewFocus(Deps{Repo: &memRepo{}, Timers: timers, Clock: movingClock{&now}, IDs: ids{}, Location: time.UTC})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
	if tm, err := u.Timer(guest); err != nil || !tm.Untouched() {
		t.Fatalf("guest read %+v %v", tm, err)
	}
	if _, err := u.ControlTimer(guest, TimerStart); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("guest start: %v", err)
	}
	if _, err := u.ConfigureTimer(guest, domain.DefaultTimerConfig); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("guest config: %v", err)
	}
	if timers.saves != 0 {
		t.Fatal("a guest stored a timer")
	}
}

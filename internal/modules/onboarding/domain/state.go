package domain

import (
	"regexp"
	"time"

	"waylight/internal/shared/kernel"
)

// stepPattern limits a wizard step id; the ids themselves belong to the
// frontend, which decides what the steps are.
var stepPattern = regexp.MustCompile(`^[a-z][a-z0-9:_-]{0,39}$`)

// State tracks the onboarding: the wizard step the account stopped at and
// whether the tour has been finished. The zero value means "not started",
// which is what a fresh installation and existing accounts start with.
type State struct {
	completedAt time.Time
	step        string
}

// RestoreState rebuilds a state from persisted data (zero time = pending,
// empty step = the wizard has not started).
func RestoreState(completedAt time.Time, step string) State {
	return State{completedAt: completedAt, step: step}
}

// Completed reports whether the tour has been finished or skipped.
func (s State) Completed() bool { return !s.completedAt.IsZero() }

// CompletedAt returns when the tour was completed (zero if pending).
func (s State) CompletedAt() time.Time { return s.completedAt }

// Step returns the wizard step the account stopped at ("" = not started).
func (s State) Step() string { return s.step }

// Complete marks the tour as done and ends the wizard. Completing twice
// keeps the first time.
func (s *State) Complete(now time.Time) {
	if s.completedAt.IsZero() {
		s.completedAt = now
	}
	s.step = ""
}

// Restart makes the wizard and the tour show again from the start.
func (s *State) Restart() {
	s.completedAt = time.Time{}
	s.step = ""
}

// Resume makes the tour pending and remembers the wizard step, so a restart
// of the tour alone can skip the wizard and an interrupted wizard continues.
func (s *State) Resume(step string) error {
	if !stepPattern.MatchString(step) {
		return kernel.NewValidationError("step", "The step is invalid.")
	}
	s.completedAt = time.Time{}
	s.step = step
	return nil
}

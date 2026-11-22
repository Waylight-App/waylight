package domain

import (
	"testing"
	"time"
)

func TestState(t *testing.T) {
	var s State
	if s.Completed() {
		t.Fatal("zero state must be pending")
	}
	first := time.Unix(100, 0)
	s.Complete(first)
	s.Complete(time.Unix(200, 0))
	if !s.Completed() || !s.CompletedAt().Equal(first) {
		t.Fatalf("unexpected %v", s.CompletedAt())
	}
	s.Restart()
	if s.Completed() {
		t.Fatal("restart must make the tour pending again")
	}
}

func TestResume(t *testing.T) {
	var s State
	if err := s.Resume("widgets"); err != nil || s.Step() != "widgets" || s.Completed() {
		t.Fatalf("resume: %v %q", err, s.Step())
	}
	if err := s.Resume("Not Valid"); err == nil {
		t.Fatal("an invalid step id must be rejected")
	}
	s.Complete(time.Unix(100, 0))
	if s.Step() != "" {
		t.Fatal("completing ends the wizard")
	}
	_ = s.Resume("done")
	s.Restart()
	if s.Step() != "" {
		t.Fatal("restart starts the wizard again")
	}
}

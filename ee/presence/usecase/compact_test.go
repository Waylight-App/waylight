package usecase

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"waylight/ee/presence/domain"
)

func TestCompact(t *testing.T) {
	store := &memStore{entries: map[string]domain.Entry{}}
	s := New(Deps{Store: store, Directory: directory{}, Absences: absences{"cleo": {Until: "2026-11-18"}},
		Clock: clock{time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)}, Location: time.UTC})
	anna := as("member", "anna")
	if err := s.Set(anna, domain.StatusOffice, ""); err != nil {
		t.Fatal(err)
	}

	state, err := s.Compact(anna, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.Title.String(); got != "1 in the office · 0 at home" {
		t.Errorf("title %q", got)
	}
	if got := state.Detail.String(); got != "1 away" {
		t.Errorf("detail %q", got)
	}
	a := state.Action
	if a == nil || a.ID != "presence.set" || a.Method != http.MethodPut || a.Path != "/api/presence/me" || len(a.Choices) != 3 {
		t.Fatalf("action %+v", a)
	}
	for _, c := range a.Choices {
		if c.Body["status"] != c.ID || c.Selected != (c.ID == "office") {
			t.Errorf("choice %+v", c)
		}
	}

	// The group narrows the counts; nobody away means no detail.
	if state, _ = s.Compact(anna, "team", true); state.Detail.Template != "" {
		t.Errorf("team detail %q", state.Detail.String())
	}
	// Read-only licence and guests get no chips.
	if state, _ = s.Compact(anna, "", false); state.Action != nil {
		t.Error("chips while read-only")
	}
	if state, _ = s.Compact(as("guest", "g"), "", true); state.Action != nil {
		t.Error("chips for a guest")
	}
	if _, err = s.Compact(anna, "nope", true); !errors.Is(err, domain.ErrNoGroup) {
		t.Errorf("unknown group: %v", err)
	}
}

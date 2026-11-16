package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/ee/presence/domain"
	"trisse-home/internal/shared/kernel"
)

type clock struct{ t time.Time }

func (c clock) Now() time.Time { return c.t }

type memStore struct {
	entries map[string]domain.Entry
	off     bool
}

func (m *memStore) Entries(context.Context) ([]domain.Entry, error) {
	out := make([]domain.Entry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out, nil
}
func (m *memStore) Save(_ context.Context, e domain.Entry) error {
	m.entries[e.AccountID] = e
	return nil
}
func (m *memStore) Delete(_ context.Context, id string) error { delete(m.entries, id); return nil }
func (m *memStore) DeleteBefore(_ context.Context, day string) error {
	for id, e := range m.entries {
		if e.Day < day {
			delete(m.entries, id)
		}
	}
	return nil
}
func (m *memStore) Enabled(context.Context) (bool, error)       { return !m.off, nil }
func (m *memStore) SetEnabled(_ context.Context, on bool) error { m.off = !on; return nil }

type directory struct{}

func (directory) People(context.Context) (map[string]string, error) {
	return map[string]string{"anna": "Anna", "ben": "Ben", "cleo": "Cleo"}, nil
}
func (directory) GroupMembers(_ context.Context, id string) ([]string, error) {
	if id == "team" {
		return []string{"anna", "ben", "gone"}, nil
	}
	return nil, domain.ErrNoGroup
}

type absences map[string]domain.Absent

func (a absences) AwayOn(context.Context, string) (map[string]domain.Absent, error) { return a, nil }

func as(role, id string) context.Context {
	return kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: id, Role: role})
}

func TestBoard(t *testing.T) {
	store := &memStore{entries: map[string]domain.Entry{}}
	s := New(Deps{Store: store, Directory: directory{}, Absences: absences{"cleo": {Until: "2026-11-18"}},
		Clock: clock{time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)}, Location: time.UTC})

	anna := as("member", "anna")
	if err := s.Set(anna, domain.StatusOffice, ""); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(as("member", "ben"), domain.StatusAway, "2026-11-17"); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(anna, "lunch", ""); !errors.Is(err, domain.ErrStatus) {
		t.Fatalf("unknown status: %v", err)
	}
	if err := s.Set(as("guest", "g"), domain.StatusOffice, ""); !errors.Is(err, domain.ErrNotSelf) {
		t.Fatalf("guest: %v", err)
	}

	b, err := s.Board(anna, "")
	if err != nil {
		t.Fatal(err)
	}
	if b.Day != "2026-11-16" || len(b.People) != 3 || b.People[0].Name != "Anna" ||
		b.Counts[domain.StatusOffice] != 1 || b.Counts[domain.StatusAway] != 2 || b.Mine.Status != domain.StatusOffice {
		t.Fatalf("board: %+v", b)
	}
	if cleo := b.People[2]; cleo.Source != domain.SourceAbsence || cleo.ReturnOn != "2026-11-19" {
		t.Fatalf("cleo: %+v", cleo)
	}

	team, err := s.Board(as("kiosk", "k"), "team")
	if err != nil || len(team.People) != 2 || team.Mine.AccountID != "" {
		t.Fatalf("team board: %+v, %v", team, err)
	}
	if _, err := s.Board(anna, "nope"); !errors.Is(err, domain.ErrNoGroup) {
		t.Fatalf("unknown group: %v", err)
	}

	if err := s.Set(anna, domain.StatusNone, ""); err != nil {
		t.Fatal(err)
	}
	if b, _ := s.Board(anna, ""); b.Mine.Status != domain.StatusNone {
		t.Fatalf("cleared: %+v", b.Mine)
	}
}

func TestSwitchAndPrune(t *testing.T) {
	store := &memStore{entries: map[string]domain.Entry{"old": {AccountID: "old", Status: domain.StatusHome, Day: "2026-11-15"}}}
	s := New(Deps{Store: store, Directory: directory{}, Clock: clock{time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)}, Location: time.UTC})

	if err := s.Prune(context.Background()); err != nil || len(store.entries) != 0 {
		t.Fatalf("prune: %v, %v", store.entries, err)
	}
	if err := s.SetEnabled(as("member", "anna"), false); !errors.Is(err, domain.ErrAdmin) {
		t.Fatalf("member switch: %v", err)
	}
	if err := s.Set(as("member", "anna"), domain.StatusHome, ""); err != nil {
		t.Fatal(err)
	}
	if err := s.SetEnabled(as("admin", "boss"), false); err != nil {
		t.Fatal(err)
	}
	if len(store.entries) != 0 {
		t.Fatalf("switching off keeps statuses: %v", store.entries)
	}
	if _, err := s.Board(as("member", "anna"), ""); !errors.Is(err, domain.ErrDisabled) {
		t.Fatalf("board while off: %v", err)
	}
	if err := s.Set(as("member", "anna"), domain.StatusHome, ""); !errors.Is(err, domain.ErrDisabled) {
		t.Fatalf("set while off: %v", err)
	}
}

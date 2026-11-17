package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/shared/kernel"
)

type memSharing struct{ tabs map[string]Tab }

func (m *memSharing) OwnTabs(context.Context) ([]Tab, error) {
	out := []Tab{}
	for _, t := range m.tabs {
		out = append(out, t)
	}
	return out, nil
}

func (m *memSharing) Share(_ context.Context, id string, v kernel.Visibility) error {
	t := m.tabs[id]
	t.Visibility = v
	m.tabs[id] = t
	return nil
}

func TestShare(t *testing.T) {
	admin := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner", Groups: []string{"office"}})
	member := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "m", Role: "member"})
	s := New(&memSharing{tabs: map[string]Tab{"t1": {ID: "t1", Name: "Tools", Visibility: kernel.Private}}}, memDefaults{})
	if _, err := s.Share(member, "t1", "everyone"); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal(err)
	}
	if _, err := s.Share(admin, "t1", "bogus"); err == nil {
		t.Fatal("an invalid visibility is refused")
	}
	v, err := s.Share(admin, "t1", "group:office")
	if err != nil || v != kernel.GroupVisibility("office") {
		t.Fatal(v, err)
	}
	if tabs, _ := s.Tabs(admin); len(tabs) != 1 || tabs[0].Visibility != v {
		t.Fatalf("%+v", tabs)
	}
	if _, err := s.Tabs(member); !errors.Is(err, kernel.ErrAdminOnly) {
		t.Fatal(err)
	}
}

type memDefaults map[string]string

func (m memDefaults) All(context.Context) (map[string]string, error) { return m, nil }
func (m memDefaults) Set(_ context.Context, g, t string) error {
	if t == "" {
		delete(m, g)
	} else {
		m[g] = t
	}
	return nil
}

func TestGroupDefault(t *testing.T) {
	admin := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner"})
	sharing := &memSharing{tabs: map[string]Tab{
		"office": {ID: "office", Visibility: kernel.GroupVisibility("g-office")},
		"mine":   {ID: "mine", Visibility: kernel.Private},
	}}
	s := New(sharing, memDefaults{})
	if err := s.SetDefault(admin, "g-office", "mine"); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatal("a private tab cannot be a group's default", err)
	}
	if err := s.SetDefault(admin, "g-office", "office"); err != nil {
		t.Fatal(err)
	}
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member", Groups: []string{"g-x", "g-office"}})
	if got := s.DefaultFor(anna); got != "office" {
		t.Fatalf("default for anna: %q", got)
	}
	if got := s.DefaultFor(kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "b", Role: "member"})); got != "" {
		t.Fatal("no groups, no default")
	}
	_ = s.SetDefault(admin, "g-office", "")
	if got := s.DefaultFor(anna); got != "" {
		t.Fatal("removed default")
	}
}

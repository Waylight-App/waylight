package usecase

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/shared/kernel"
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
	s := New(&memSharing{tabs: map[string]Tab{"t1": {ID: "t1", Name: "Tools", Visibility: kernel.Private}}})
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

package usecase

import (
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	up := func(name string) ServiceView { return ServiceView{Name: name, State: "online"} }
	down := func(name string) ServiceView { return ServiceView{Name: name, State: "offline"} }

	if s := CompactState(nil, now); s.Title.String() != "No services yet" || s.Level != kernel.CompactNone || !s.Empty {
		t.Fatalf("empty: %+v", s)
	}
	s := CompactState([]ServiceView{up("Wiki"), up("Git"), {Name: "Gone", State: "offline", Hidden: true}}, now)
	if s.Title.String() != "2 of 2 running" || s.Level != kernel.CompactOK || s.Empty || s.Detail.Template != "" || *s.Count != 0 {
		t.Fatalf("all up: %+v", s)
	}
	s = CompactState([]ServiceView{up("Wiki"), down("Nextcloud"), {Name: "New", State: "unknown"}}, now)
	if s.Title.String() != "1 of 3 running" || s.Level != kernel.CompactProblem || s.Detail.String() != "Nextcloud is down" || *s.Count != 1 {
		t.Fatalf("one down: %+v", s)
	}
	s = CompactState([]ServiceView{down("A"), down("B")}, now)
	if s.Detail.String() != "2 services down" || *s.Count != 2 {
		t.Fatalf("two down: %+v", s)
	}
	// Not checked yet: neither good nor bad.
	if s = CompactState([]ServiceView{{Name: "New", State: "unknown"}}, now); s.Level != kernel.CompactNone {
		t.Fatalf("unchecked: %+v", s)
	}
}

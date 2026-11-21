package usecase

import (
	"testing"

	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	v := ListView{Peers: make([]PeerView, 3), Online: 2}
	s := CompactState(v)
	if s.Title.String() != "2 of 3 peers connected" || s.Detail.Template != "" || s.Level != kernel.CompactNone || *s.Count != 2 {
		t.Fatalf("fresh: %+v", s)
	}
	v.Stale = true
	if s := CompactState(v); s.Level != kernel.CompactAttention || s.Detail.String() != "The export is out of date" {
		t.Fatalf("stale: %+v", s)
	}
	if s := CompactState(ListView{}); s.Title.String() != "No peers" || !s.Empty {
		t.Fatalf("empty: %+v", s)
	}
	if s := CompactState(ListView{Stale: true}); s.Empty {
		t.Fatalf("a stale export stays visible: %+v", s)
	}
}

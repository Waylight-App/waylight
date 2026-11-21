package usecase

import (
	"testing"

	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	s := CompactState(ListResponse{Total: 12, Running: 11, Stopped: 1})
	if s.Title.String() != "11 of 12 running" || s.Detail.String() != "1 stopped" || s.Level != kernel.CompactNone {
		t.Fatalf("stopped: %+v", s)
	}
	if s := CompactState(ListResponse{Total: 3, Running: 3, Unhealthy: 1}); s.Level != kernel.CompactAttention || s.Detail.String() != "1 unhealthy" {
		t.Fatalf("unhealthy: %+v", s)
	}
	if s := CompactState(ListResponse{Total: 3, Running: 3, Error: "refused"}); s.Level != kernel.CompactAttention || s.Detail.String() != "Docker is not reachable" {
		t.Fatalf("unreachable: %+v", s)
	}
	if s := CompactState(ListResponse{Total: 2, Running: 2}); s.Detail.Template != "" || s.Empty {
		t.Fatalf("all running: %+v", s)
	}
	if s := CompactState(ListResponse{}); !s.Empty {
		t.Fatalf("no containers: %+v", s)
	}
	if s := CompactState(ListResponse{Error: "refused"}); s.Empty {
		t.Fatalf("unreachable stays visible: %+v", s)
	}
}

package usecase

import (
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	v := StatsView{Queries: 1000, Blocked: 123, BlockedPercent: 12.3, ProtectionEnabled: true}
	s := CompactState(v)
	if s.Title.String() != "12 % blocked" || s.Detail.String() != "123 of 1000 queries" || s.Level != kernel.CompactNone {
		t.Fatalf("enabled: %+v", s)
	}
	v.ProtectionEnabled = false
	if s := CompactState(v); s.Level != kernel.CompactAttention || s.Detail.String() != "Blocking is off" {
		t.Fatalf("off: %+v", s)
	}
	until := time.Date(2026, 11, 21, 12, 5, 0, 0, time.UTC)
	v.PausedUntil = &until
	if s := CompactState(v); s.Level != kernel.CompactNone || s.Detail.String() != "Blocking paused" {
		t.Fatalf("paused: %+v", s)
	}
	v.Error = "timeout"
	if s := CompactState(v); s.Level != kernel.CompactAttention || s.Detail.String() != "The DNS filter is not reachable" {
		t.Fatalf("unreachable: %+v", s)
	}
}

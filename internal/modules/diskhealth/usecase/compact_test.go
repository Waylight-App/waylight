package usecase

import (
	"testing"
	"time"

	"waylight/internal/modules/diskhealth/domain"
	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	disk := func(dev string, s domain.Status) domain.Disk {
		return domain.Disk{Report: domain.Report{Device: dev}, Status: s}
	}
	if s := CompactState(nil, now); s.Level != kernel.CompactNone || s.Title.String() != "No drive reports yet" || !s.Empty {
		t.Fatalf("empty: %+v", s)
	}
	s := CompactState([]domain.Disk{disk("/dev/sda", domain.StatusOK), disk("/dev/sdb", domain.StatusOK)}, now)
	if s.Level != kernel.CompactOK || s.Title.String() != "2 drives healthy" || s.Empty {
		t.Fatalf("healthy: %+v", s)
	}
	s = CompactState([]domain.Disk{disk("/dev/sda", domain.StatusWarning), disk("/dev/sdb", domain.StatusOK)}, now)
	if s.Level != kernel.CompactAttention || s.Detail.String() != "/dev/sda has warnings" {
		t.Fatalf("warning: %+v", s)
	}
	s = CompactState([]domain.Disk{disk("/dev/sda", domain.StatusWarning), disk("/dev/sdb", domain.StatusCritical)}, now)
	if s.Level != kernel.CompactProblem || s.Title.String() != "1 of 2 drives failing" || *s.Count != 1 {
		t.Fatalf("critical: %+v", s)
	}
	s = CompactState([]domain.Disk{disk("/dev/sda", domain.StatusUnknown), disk("/dev/sdb", domain.StatusOK)}, now)
	if s.Level != kernel.CompactNone || s.Title.String() != "1 of 2 drives healthy" {
		t.Fatalf("stale: %+v", s)
	}
}

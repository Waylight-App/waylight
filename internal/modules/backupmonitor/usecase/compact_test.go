package usecase

import (
	"testing"
	"time"

	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	at3 := time.Date(2026, 11, 21, 3, 0, 0, 0, time.UTC)
	if s := CompactState(nil, now, time.UTC); s.Title.String() != "No backups yet" || s.Level != kernel.CompactNone {
		t.Fatalf("empty: %+v", s)
	}
	s := CompactState([]SourceView{{Name: "NAS", State: "ok", LastSuccessAt: at3}}, now, time.UTC)
	if s.Title.String() != "Backup OK" || s.Level != kernel.CompactOK || s.Detail.String() != "Last success 2026-11-21 03:00" {
		t.Fatalf("ok: %+v", s)
	}
	s = CompactState([]SourceView{{Name: "NAS", State: "ok"}, {Name: "Photos", State: "overdue"}}, now, time.UTC)
	if s.Title.String() != "1 of 2 overdue" || s.Level != kernel.CompactAttention || s.Detail.String() != "Photos is overdue" {
		t.Fatalf("overdue: %+v", s)
	}
	s = CompactState([]SourceView{{Name: "NAS", State: "failed"}, {Name: "Photos", State: "overdue"}}, now, time.UTC)
	if s.Level != kernel.CompactProblem || s.Detail.String() != "NAS failed" || *s.Count != 1 {
		t.Fatalf("failed wins: %+v", s)
	}
	if s = CompactState([]SourceView{{Name: "New", State: "waiting"}}, now, time.UTC); s.Title.String() != "Waiting for the first run" || s.Level != kernel.CompactNone {
		t.Fatalf("waiting: %+v", s)
	}
}

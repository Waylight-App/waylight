package usecase

import (
	"testing"

	"waylight/internal/modules/speedtest/domain"
	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	m := &domain.Measurement{DownloadBps: 250.4e6, UploadBps: 39.6e6, PingMs: 12.3}
	s := CompactState(ResultsView{Latest: m})
	if s.Title.String() != "250 ↓ · 40 ↑ Mbit/s" || s.Detail.String() != "Ping 12 ms" || s.Level != kernel.CompactNone {
		t.Fatalf("%+v", s)
	}
	if s := CompactState(ResultsView{Latest: m, Error: "timeout"}); s.Level != kernel.CompactAttention || s.Detail.String() != "Speedtest Tracker is not reachable" {
		t.Fatalf("error: %+v", s)
	}
	if s := CompactState(ResultsView{}); s.Title.String() != "No results yet" || !s.Empty {
		t.Fatalf("empty: %+v", s)
	}
	if s := CompactState(ResultsView{Error: "timeout"}); s.Empty || s.Level != kernel.CompactAttention {
		t.Fatalf("unreachable without results stays visible: %+v", s)
	}
}

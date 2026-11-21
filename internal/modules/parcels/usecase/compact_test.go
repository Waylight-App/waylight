package usecase

import (
	"testing"
	"time"

	"waylight/internal/modules/parcels/domain"
	"waylight/internal/shared/kernel"
)

func TestCompactState(t *testing.T) {
	now := time.Date(2026, 11, 21, 9, 0, 0, 0, time.UTC)
	transit := domain.Parcel{Status: domain.Status{Code: domain.StatusTransit}}
	arriving := domain.Parcel{Status: domain.Status{Code: domain.StatusTransit}, EstimatedDelivery: now.Add(5 * time.Hour)}
	delivered := domain.Parcel{Status: domain.Status{Code: domain.StatusDelivered, At: now}}

	s := CompactState([]domain.Parcel{transit, arriving, delivered}, now, time.UTC)
	if s.Title.String() != "2 on the way" || s.Detail.String() != "1 arriving today" || *s.Count != 2 || s.Level != kernel.CompactNone {
		t.Fatalf("%+v", s)
	}
	failed := domain.Parcel{Status: domain.Status{Code: domain.StatusFailure}}
	if s.Empty {
		t.Fatalf("parcels on the way are not empty: %+v", s)
	}
	if s := CompactState([]domain.Parcel{arriving, failed}, now, time.UTC); s.Level != kernel.CompactAttention || s.Detail.String() != "1 with a problem" {
		t.Fatalf("failure: %+v", s)
	}
	if s := CompactState([]domain.Parcel{delivered}, now, time.UTC); s.Title.String() != "No parcels on the way" || *s.Count != 0 || !s.Empty {
		t.Fatalf("none: %+v", s)
	}
}

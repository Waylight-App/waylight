package contracts

import (
	"context"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/contracts/adapter/gateway"
	"waylight/internal/modules/contracts/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return strconv.Itoa(s.n) }

func TestCompactState(t *testing.T) {
	m := New(dbtest.Open(t, "contracts", gateway.Migrations()), fixedClock{time.Date(2026, 12, 20, 9, 0, 0, 0, time.UTC)}, &seqIDs{}, nil, time.UTC)
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if s, _ := m.CompactState(ctx); !s.Empty || s.Title.String() != "No deadlines ahead" {
		t.Fatalf("empty: %+v", s)
	}
	for _, in := range []domain.Input{
		{Name: "Gym", Period: "monthly", TermEnd: "2027-06-30", NoticeDays: 30, RenewMonths: 12, RemindDays: 14},
		{Name: "Mobile", Period: "monthly", TermEnd: "2027-03-31", NoticeDays: 90, RenewMonths: 12, RemindDays: 14},
	} {
		if _, err := m.uc.Create(ctx, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	s, err := m.CompactState(ctx)
	if err != nil || s.Title.String() != "Mobile: cancel by 2026-12-31" || s.Level != kernel.CompactAttention || *s.Count != 1 {
		t.Fatalf("%+v %v", s, err)
	}
}

package board

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/board/adapter/gateway"
	"waylight/internal/modules/board/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return strconv.Itoa(s.n) }

func TestCompactState(t *testing.T) {
	m := New(dbtest.Open(t, "board", gateway.Migrations()), fixedClock{time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC)}, &seqIDs{}, time.UTC)
	ctx := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	if s, _ := m.CompactState(ctx); !s.Empty || s.Title.String() != "The board is empty" {
		t.Fatalf("empty: %+v", s)
	}
	for _, in := range []domain.Input{{Text: "Short"}, {Text: strings.Repeat("Pinned ", 20) + "\nsecond line", Pinned: true}} {
		if _, err := m.uc.Create(ctx, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	s, err := m.CompactState(ctx)
	detail := s.Detail.String()
	if err != nil || s.Title.String() != "2 notes" || !strings.HasPrefix(detail, "Pinned") || !strings.HasSuffix(detail, "…") || len([]rune(detail)) != compactTextLength {
		t.Fatalf("%+v %q %v", s, detail, err)
	}
}

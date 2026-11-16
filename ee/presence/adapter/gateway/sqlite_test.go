package gateway

import (
	"context"
	"testing"

	"trisse-home/ee/presence/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestStore(t *testing.T) {
	ctx := context.Background()
	s := NewStore(dbtest.Open(t, "presence", Migrations()))

	if on, err := s.Enabled(ctx); err != nil || !on {
		t.Fatalf("Enabled() = %v, %v; want on by default", on, err)
	}
	if err := s.SetEnabled(ctx, false); err != nil {
		t.Fatal(err)
	}
	if on, _ := s.Enabled(ctx); on {
		t.Fatal("still on after SetEnabled(false)")
	}

	for _, e := range []domain.Entry{
		{AccountID: "a", Status: domain.StatusOffice, Day: "2026-11-15"},
		{AccountID: "b", Status: domain.StatusAway, Day: "2026-11-16", ReturnOn: "2026-11-18"},
		{AccountID: "a", Status: domain.StatusHome, Day: "2026-11-16"},
	} {
		if err := s.Save(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	list, err := s.Entries(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("Entries() = %v, %v; want one row per account", list, err)
	}

	if err := s.Save(ctx, domain.Entry{AccountID: "c", Status: domain.StatusOffice, Day: "2026-11-15"}); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteBefore(ctx, "2026-11-16"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, "missing"); err != nil {
		t.Fatalf("Delete(missing) = %v", err)
	}
	list, _ = s.Entries(ctx)
	if len(list) != 1 || list[0] != (domain.Entry{AccountID: "b", Status: domain.StatusAway, Day: "2026-11-16", ReturnOn: "2026-11-18"}) {
		t.Fatalf("after prune: %+v", list)
	}
}

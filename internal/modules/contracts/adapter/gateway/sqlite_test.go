package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/contracts/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

func contract(t *testing.T, id string, v kernel.Visibility) domain.Contract {
	t.Helper()
	c, err := domain.New(id, domain.Input{Name: "Contract " + id, Provider: "ACME", CostCents: 1999, Period: "quarterly", TermEnd: "2027-03-31",
		NoticeDays: 90, RenewMonths: 12, RemindDays: 14, Note: "note"}, kernel.Access{Visibility: v}, time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// TestOwnership proves ADR 0018 for contracts, and that a change by the
// owner keeps the reminder bookkeeping.
func TestOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "contracts", Migrations()))
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	shared, private := contract(t, "shared", kernel.Everyone), contract(t, "private", kernel.Private)
	for _, c := range []domain.Contract{shared, private} {
		if err := repo.Save(anna, c); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := repo.List(ben)
	if len(list) != 1 || list[0].ID != "shared" || list[0].Period != domain.Quarterly || list[0].TermEndDate() != "2027-03-31" || list[0].Note != "note" {
		t.Fatalf("ben sees %+v", list)
	}
	if _, err := repo.Get(ben, "private"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben read anna's private contract: %v", err)
	}
	shared.Name = "Mine"
	if err := repo.Save(ben, shared); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's contract: %v", err)
	}
	if err := repo.Delete(ben, "shared"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted anna's contract: %v", err)
	}
	if err := repo.MarkReminded(context.Background(), "private", "2026-12-31"); err != nil {
		t.Fatal(err)
	}
	private.Name = "Renamed"
	if err := repo.Save(anna, private); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(anna, "private"); got.Name != "Renamed" || got.RemindedFor != "2026-12-31" {
		t.Fatalf("%+v", got)
	}
}

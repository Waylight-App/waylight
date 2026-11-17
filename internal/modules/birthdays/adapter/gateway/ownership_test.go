package gateway

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/birthdays/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestBirthdayOwnership proves ADR 0018 for birthdays, including the
// address book sync, which only replaces the principal's own entries.
func TestBirthdayOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "birthdays", Migrations()), prefixCipher{})
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})

	if err := repo.Save(anna, domain.Birthday{ID: "b-anna", Name: "Secret friend", Month: 3, Day: 4, Source: "manual"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(admin, domain.Birthday{ID: "b-admin", Name: "Grandma", Month: 5, Day: 6, Source: "manual"}); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(ben); len(list) != 1 || list[0].ID != "b-admin" {
		t.Fatalf("ben sees %v", list)
	}
	if err := repo.Save(ben, domain.Birthday{ID: "b-admin", Name: "Mine", Month: 1, Day: 1}); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed the admin's birthday: %v", err)
	}
	if err := repo.Delete(ben, "b-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's birthday: %v", err)
	}
	if err := repo.Delete(ben, "b-anna"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an invisible birthday must look missing: %v", err)
	}
	// Ben's sync of the "manual" source leaves the others' entries alone.
	if err := repo.ReplaceSource(ben, "manual", nil); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(context.Background()); len(list) != 2 {
		t.Fatalf("ben's sync removed others' birthdays: %v", list)
	}
}

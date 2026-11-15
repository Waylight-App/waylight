package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/parcels/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestParcelOwnership proves ADR 0018 for parcels.
func TestParcelOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "parcels", Migrations()), prefixCipher{})
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})

	if err := repo.Save(anna, domain.Parcel{ID: "p-anna", TrackingNumber: "00340434161094042557", Carrier: domain.CarrierDHL, Label: "Gift", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	family := domain.Parcel{ID: "p-admin", TrackingNumber: "00340434161094042558", Carrier: domain.CarrierDHL, Label: "Sofa", CreatedAt: now}
	if err := repo.Save(admin, family); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(ben); len(list) != 1 || list[0].ID != "p-admin" {
		t.Fatalf("ben sees %v", list)
	}
	if _, err := repo.Get(admin, "p-anna"); !errors.Is(err, domain.ErrParcelNotFound) {
		t.Fatalf("the admin saw anna's gift: %v", err)
	}
	family.Label = "Mine"
	if err := repo.Save(ben, family); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben relabelled the admin's parcel: %v", err)
	}
	if err := repo.Delete(ben, "p-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's parcel: %v", err)
	}
	// The tracking job (system) updates everyone's parcels.
	family.Label = "Sofa"
	family.Error = "temporarily unavailable"
	if err := repo.Save(context.Background(), family); err != nil {
		t.Fatal(err)
	}
}

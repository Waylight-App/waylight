package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/photos/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestPhotoOwnership proves ADR 0018 for photos: private photos stay
// private, shared ones are visible but only their owner removes or moves
// them.
func TestPhotoOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "photos", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	for ctx, id := range map[context.Context]string{admin: "p-admin", anna: "p-anna"} {
		if _, err := repo.Add(ctx, domain.Photo{ID: id, Album: "Holiday", Width: 1, Height: 1, Size: 10, CreatedAt: now}, []byte{1}, []byte{2}); err != nil {
			t.Fatal(err)
		}
	}
	if list, _ := repo.List(ben, "Holiday"); len(list) != 1 || list[0].ID != "p-admin" {
		t.Fatalf("ben sees %v", list)
	}
	if albums, _ := repo.Albums(ben); len(albums) != 1 || albums[0].Count != 1 {
		t.Fatalf("ben's albums %v", albums)
	}
	if _, err := repo.Image(ben, "p-anna", true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ben loaded anna's photo: %v", err)
	}
	if err := repo.Delete(ben, "p-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's photo: %v", err)
	}
	if err := repo.Delete(ben, "p-anna"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("an invisible photo must look missing: %v", err)
	}
	if err := repo.Reorder(ben, "Holiday", []string{"x", "p-admin"}); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(admin, "Holiday"); list[0].Position != 0 {
		t.Fatalf("ben moved the admin's photo to %d", list[0].Position)
	}
	if u, _ := repo.Usage(ben); u.Count != 2 {
		t.Fatalf("usage covers the instance, got %d", u.Count)
	}
}

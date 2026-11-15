package gateway

import (
	"context"
	"testing"

	"trisse-home/internal/modules/releases/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestFollowedArtistsArePersonal: two accounts can follow the same artist,
// each sees its own list, and unfollowing affects only the own list.
func TestFollowedArtistsArePersonal(t *testing.T) {
	repo := NewArtistRepository(dbtest.Open(t, "releases", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
	same := domain.Artist{ID: "a74b1b7f-71a5-4011-9441-d0b5e4122711", Name: "Radiohead"}

	for _, ctx := range []context.Context{admin, anna} {
		if err := repo.Add(ctx, same); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.Add(anna, domain.Artist{ID: "b10bbbfc-cf9e-42e0-be17-e2c3e1d2600d", Name: "The Beatles"}); err != nil {
		t.Fatal(err)
	}
	count := func(ctx context.Context) int {
		list, err := repo.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return len(list)
	}
	if count(admin) != 1 || count(anna) != 2 || count(guest) != 1 || count(context.Background()) != 3 {
		t.Fatalf("admin %d, anna %d, guest %d, system %d", count(admin), count(anna), count(guest), count(context.Background()))
	}
	if err := repo.Remove(anna, same.ID); err != nil {
		t.Fatal(err)
	}
	if count(admin) != 1 || count(anna) != 1 {
		t.Fatalf("anna's unfollow changed the admin's list: admin %d, anna %d", count(admin), count(anna))
	}
}

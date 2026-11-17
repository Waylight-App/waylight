package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/bookmarks/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestFolderSharedWithGroup: a folder the admin shares with a group is that
// group's link directory (#298). Its bookmarks, also ones added later, are
// visible to the group's members, read-only, and to nobody else.
func TestFolderSharedWithGroup(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "bookmarks", Migrations()))
	now := time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner", PrivateByDefault: true})
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member", Groups: []string{"office"}})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})

	tools, _ := domain.NewFolder("f-tools", "Office tools", 0)
	if err := repo.SaveFolder(admin, tools); err != nil {
		t.Fatal(err)
	}
	mk := func(id, folder string) {
		b, _ := domain.NewBookmark(id, domain.BookmarkInput{Title: id, URL: "https://example.org/" + id, FolderID: folder}, now)
		if err := repo.SaveBookmark(admin, b); err != nil {
			t.Fatal(err)
		}
	}
	mk("b-before", "f-tools")
	mk("b-personal", "")
	if list, _ := repo.ListBookmarks(anna); len(list) != 0 {
		t.Fatalf("private folder leaked: %v", list)
	}

	if err := repo.SetFolderVisibility(anna, "f-tools", kernel.GroupVisibility("office")); err == nil {
		t.Fatal("only the owner shares a folder")
	}
	if err := repo.SetFolderVisibility(admin, "f-tools", kernel.GroupVisibility("office")); err != nil {
		t.Fatal(err)
	}
	mk("b-after", "f-tools")

	list, _ := repo.ListBookmarks(anna)
	if len(list) != 2 {
		t.Fatalf("anna should see both bookmarks of the shared folder: %v", list)
	}
	for _, b := range list {
		if b.FolderID() != "f-tools" || b.Access().Mine(kernel.PrincipalFrom(anna)) {
			t.Fatalf("unexpected %+v", b)
		}
	}
	if folders, _ := repo.ListFolders(anna); len(folders) != 1 || folders[0].Access().Visibility != kernel.GroupVisibility("office") {
		t.Fatalf("anna's folders: %+v", folders)
	}
	if list, _ := repo.ListBookmarks(ben); len(list) != 0 {
		t.Fatalf("ben is not in the group: %v", list)
	}
	if err := repo.SaveBookmark(anna, list[0]); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("shared links are read-only: %v", err)
	}

	if err := repo.SetFolderVisibility(admin, "f-tools", kernel.Private); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.ListBookmarks(anna); len(list) != 0 {
		t.Fatalf("unsharing the folder hides its bookmarks again: %v", list)
	}
}

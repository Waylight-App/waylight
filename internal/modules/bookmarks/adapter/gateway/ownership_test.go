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

// TestBookmarkOwnership proves ADR 0018 for bookmarks, folders, icons and
// link checks, and that config as code only replaces the admin's.
func TestBookmarkOwnership(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "bookmarks", Migrations()))
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	system := context.Background()

	work, _ := domain.NewFolder("f-admin", "Work", 0)
	if err := repo.SaveFolder(admin, work); err != nil {
		t.Fatal(err)
	}
	private, _ := domain.NewFolder("f-anna", "Private", 0)
	if err := repo.SaveFolder(anna, private); err != nil {
		t.Fatal(err)
	}
	mk := func(ctx context.Context, id, folder string) error {
		b, err := domain.NewBookmark(id, domain.BookmarkInput{Title: id, URL: "https://example.org/" + id, FolderID: folder}, now)
		if err != nil {
			t.Fatal(err)
		}
		return repo.SaveBookmark(ctx, b)
	}
	if err := mk(admin, "b-admin", "f-admin"); err != nil {
		t.Fatal(err)
	}
	if err := mk(anna, "b-anna", "f-anna"); err != nil {
		t.Fatal(err)
	}
	// Filing into someone else's (visible) folder is refused.
	if err := mk(ben, "b-ben", "f-admin"); !errors.Is(err, domain.ErrFolderNotFound) {
		t.Fatalf("ben filed into the admin's folder: %v", err)
	}
	_ = repo.SaveIcon(system, "b-anna", domain.Icon{ContentType: "image/png", Data: []byte{1}})

	if list, _ := repo.ListBookmarks(ben); len(list) != 1 || list[0].ID() != "b-admin" {
		t.Fatalf("ben sees %v", list)
	}
	if folders, _ := repo.ListFolders(ben); len(folders) != 1 {
		t.Fatalf("ben sees folders %v", folders)
	}
	if _, err := repo.Icon(ben, "b-anna"); !errors.Is(err, domain.ErrIconNotFound) {
		t.Fatalf("ben read the icon of anna's bookmark: %v", err)
	}
	if ids, _ := repo.IconIDs(ben); ids["b-anna"] {
		t.Fatal("ben learned that anna's bookmark has an icon")
	}
	if _, err := repo.Icon(anna, "b-anna"); err != nil {
		t.Fatal(err)
	}

	stolen, _ := repo.GetBookmark(ben, "b-admin")
	if err := repo.SaveBookmark(ben, stolen); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed the admin's bookmark: %v", err)
	}
	if err := repo.DeleteBookmark(ben, "b-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's bookmark: %v", err)
	}
	if err := repo.SaveFolder(ben, work); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben renamed the admin's folder: %v", err)
	}
	if err := repo.DeleteFolder(ben, "f-admin"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted the admin's folder: %v", err)
	}

	// Config as code replaces the admin's bookmarks only.
	if err := repo.ReplaceAll(admin, nil, nil); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.ListBookmarks(system); len(list) != 1 || list[0].ID() != "b-anna" {
		t.Fatalf("after replace: %v", list)
	}
	taken := domain.RestoreFolder("f-anna", "Taken", 0)
	if err := repo.ReplaceAll(admin, []domain.Folder{taken}, nil); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("replace took over anna's folder: %v", err)
	}
}

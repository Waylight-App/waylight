package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules/documents/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

func document(t *testing.T, id string, v kernel.Visibility) domain.Document {
	t.Helper()
	d, err := domain.New(id, domain.Input{Title: "Document " + id, PurchasedOn: "2026-01-02", WarrantyEnd: "2028-01-02", Note: "note"},
		kernel.Access{Visibility: v}, time.Date(2026, 11, 23, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestOwnershipAndFiles proves ADR 0018 for documents and their files, and
// that a file leaves with its document.
func TestOwnershipAndFiles(t *testing.T) {
	db := dbtest.Open(t, "documents", Migrations())
	repo := NewRepository(db)
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	shared, private := document(t, "shared", kernel.Everyone), document(t, "private", kernel.Private)
	file := domain.File{Name: "receipt.pdf", ContentType: "application/pdf", Size: 4}
	for _, d := range []domain.Document{shared, private} {
		if err := repo.Save(anna, d); err != nil {
			t.Fatal(err)
		}
		if err := repo.SaveFile(anna, d.ID, file, []byte("%PDF")); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := repo.List(ben)
	if len(list) != 1 || list[0].ID != "shared" || domain.Day(list[0].WarrantyEnd) != "2028-01-02" || list[0].File != file {
		t.Fatalf("ben sees %+v", list)
	}
	if data, err := repo.File(ben, "shared"); err != nil || string(data) != "%PDF" {
		t.Fatalf("ben downloads the shared file: %q %v", data, err)
	}
	if _, err := repo.File(ben, "private"); !errors.Is(err, domain.ErrNoFile) {
		t.Fatalf("ben read anna's private file: %v", err)
	}
	if err := repo.SaveFile(ben, "shared", file, []byte("x")); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben replaced anna's file: %v", err)
	}
	shared.Title = "Mine"
	if err := repo.Save(ben, shared); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben changed anna's document: %v", err)
	}
	if err := repo.Delete(ben, "shared"); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben deleted anna's document: %v", err)
	}
	// Changing the fields keeps the file; a new file replaces the old one.
	private.Title = "Renamed"
	if err := repo.Save(anna, private); err != nil {
		t.Fatal(err)
	}
	if err := repo.SaveFile(anna, "private", domain.File{Name: "b.png", ContentType: "image/png", Size: 2}, []byte("ab")); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(anna, "private"); got.Title != "Renamed" || got.File.Name != "b.png" {
		t.Fatalf("%+v", got)
	}
	if err := repo.Delete(anna, "private"); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM documents_files`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("files left: %d %v", n, err)
	}
}

package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/phonelist/domain"
	"trisse-home/internal/modules/phonelist/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "phonelist", Migrations()), plainCipher{})
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	ben, _ := domain.NewEntry("2", domain.EntryInput{Name: "ben", Email: "ben@example.org"}, at)
	anna, _ := domain.NewEntry("1", domain.EntryInput{Name: "Anna", Role: "Reception", Phone: "+49 1", Team: "Front"}, at)
	for _, e := range []domain.Entry{ben, anna} {
		if err := repo.Save(ctx, e); err != nil {
			t.Fatal(err)
		}
	}
	list, err := repo.List(ctx)
	if err != nil || len(list) != 2 || list[0].Name != "Anna" || list[0].Team != "Front" || !list[0].CreatedAt.Equal(at) {
		t.Fatalf("%+v %v", list, err)
	}
	anna.Phone = "+49 2"
	_ = repo.Save(ctx, anna)
	if got, _ := repo.Get(ctx, "1"); got.Phone != "+49 2" {
		t.Fatalf("update: %+v", got)
	}
	_ = repo.Delete(ctx, "1")
	if _, err := repo.Get(ctx, "1"); !errors.Is(err, domain.ErrEntryNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

// plainCipher marks values instead of encrypting them.
type plainCipher struct{}

func (plainCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (plainCipher) Decrypt(s string) (string, error) { return s[len("enc:"):], nil }

func TestReplaceSourceAndSettings(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "phonelist", Migrations()), plainCipher{})
	at := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	manual, _ := domain.NewEntry("m", domain.EntryInput{Name: "Desk", Phone: "1"}, at)
	_ = repo.Save(ctx, manual)
	book, _ := domain.FromCard(domain.Card{ID: "cd-1", EntryInput: domain.EntryInput{Name: "Anna", Phone: "2"}}, at)
	if err := repo.ReplaceSource(ctx, domain.SourceCardDAV, []domain.Entry{book}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceSource(ctx, domain.SourceCardDAV, []domain.Entry{book}); err != nil {
		t.Fatal("a second sync with the same IDs", err)
	}
	list, _ := repo.List(ctx)
	if len(list) != 2 || list[0].Source != domain.SourceCardDAV || list[1].Source != domain.SourceManual {
		t.Fatalf("%+v", list)
	}
	_ = repo.ReplaceSource(ctx, domain.SourceCardDAV, nil)
	if list, _ := repo.List(ctx); len(list) != 1 || list[0].ID != "m" {
		t.Fatalf("manual entries survive a sync: %+v", list)
	}

	if s, err := repo.Settings(ctx); err != nil || s.CardDAVURL != "" {
		t.Fatal(s, err)
	}
	in := usecase.Settings{CardDAVURL: "https://dav.example/book/", Username: "u", Password: "secret", LastSync: at, LastError: ""}
	if err := repo.SaveSettings(ctx, in); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Settings(ctx)
	if err != nil || got.Password != "secret" || !got.LastSync.Equal(at) {
		t.Fatalf("%+v %v", got, err)
	}
}

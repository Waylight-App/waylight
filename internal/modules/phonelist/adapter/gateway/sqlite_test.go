package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"trisse-home/internal/modules/phonelist/domain"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "phonelist", Migrations()))
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

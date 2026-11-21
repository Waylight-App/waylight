package gateway

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/connections/domain"
	"waylight/internal/platform/database/dbtest"
)

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + strings.ToUpper(s), nil }
func (prefixCipher) Decrypt(s string) (string, error) {
	if !strings.HasPrefix(s, "enc:") {
		return "", errors.New("bad ciphertext")
	}
	return strings.ToLower(strings.TrimPrefix(s, "enc:")), nil
}

func TestRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "connections", Migrations())
	repo := NewRepository(db, prefixCipher{})
	created := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	c, _ := domain.NewConnection("1", "immich", "Home", "http://h:2283", domain.Credential{"apikey": "secret"}, created)
	if err := repo.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRowContext(ctx, `SELECT credential_encrypted FROM connections_connections WHERE id = '1'`).Scan(&stored)
	if strings.Contains(stored, "secret") {
		t.Fatalf("credential stored in plain text: %q", stored)
	}
	got, err := repo.Get(ctx, "1")
	if err != nil || got.Credential["apikey"] != "secret" || got.Name != "Home" || !got.CreatedAt.Equal(created) {
		t.Fatalf("%+v %v", got, err)
	}
	dup, _ := domain.NewConnection("2", "immich", "home", "http://x", nil, created)
	if err := repo.Save(ctx, dup); err == nil {
		t.Fatal("the unique index must refuse the same name of a service")
	}
	_ = repo.Delete(ctx, "1")
	if _, err := repo.Get(ctx, "1"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

// TestUnreadableCredential keeps the list usable after SECRET_KEY changed.
func TestUnreadableCredential(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "connections", Migrations())
	_, _ = db.ExecContext(ctx, `INSERT INTO connections_connections VALUES ('1', 'immich', 'Home', 'http://h', 'garbage', '2026-11-21T12:00:00.000000000Z')`)
	list, err := NewRepository(db, prefixCipher{}).List(ctx)
	if err != nil || len(list) != 1 || len(list[0].Credential) != 0 {
		t.Fatalf("%+v %v", list, err)
	}
}

func TestReplaceAllSwapsNames(t *testing.T) {
	ctx := context.Background()
	repo := NewRepository(dbtest.Open(t, "connections", Migrations()), prefixCipher{})
	now := time.Now()
	a, _ := domain.NewConnection("a", "immich", "A", "http://a", nil, now)
	b, _ := domain.NewConnection("b", "immich", "B", "http://b", nil, now)
	_ = repo.Save(ctx, a)
	_ = repo.Save(ctx, b)
	a.Name, b.Name = "B", "A"
	if err := repo.ReplaceAll(ctx, []domain.Connection{a, b}); err != nil {
		t.Fatal(err)
	}
	list, _ := repo.List(ctx)
	if len(list) != 2 || list[0].ID != "b" || list[0].Name != "A" {
		t.Fatalf("%+v", list)
	}
}

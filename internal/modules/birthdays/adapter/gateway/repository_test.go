package gateway

import (
	"context"
	"errors"
	"strings"
	"testing"

	"waylight/internal/modules/birthdays/domain"
	"waylight/internal/modules/birthdays/usecase"
	"waylight/internal/platform/database/dbtest"
)

type prefixCipher struct{}

func (prefixCipher) Encrypt(s string) (string, error) { return "enc:" + s, nil }
func (prefixCipher) Decrypt(s string) (string, error) { return strings.TrimPrefix(s, "enc:"), nil }

func TestRepository(t *testing.T) {
	ctx := context.Background()
	r := NewRepository(dbtest.Open(t, "birthdays", Migrations()), prefixCipher{})
	if s, err := r.Settings(ctx); err != nil || s.Reminder.At != "08:00" {
		t.Fatal(s, err)
	}
	_ = r.Save(ctx, domain.Birthday{ID: "a", Name: "Anna", Month: 5, Day: 17, Year: 1990, Source: domain.SourceManual})
	_ = r.ReplaceSource(ctx, domain.SourceCardDAV, []domain.Birthday{{ID: "c1", Name: "Carla", Month: 1, Day: 2, Source: domain.SourceCardDAV}})
	_ = r.ReplaceSource(ctx, domain.SourceCardDAV, []domain.Birthday{{ID: "c2", Name: "Dora", Month: 3, Day: 4, Source: domain.SourceCardDAV}})
	list, _ := r.List(ctx)
	if len(list) != 2 || list[0].Name != "Dora" || list[1].Name != "Anna" || list[1].Year != 1990 {
		t.Fatalf("%+v", list)
	}
	if err := r.Delete(ctx, "x"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
	s := usecase.Settings{Reminder: domain.Reminder{Enabled: true, At: "07:30"}, LastReminded: "2026-11-14", CardDAVURL: "https://x/", Username: "u", Password: "pw"}
	_ = r.SaveSettings(ctx, s)
	var enc string
	_ = r.db.QueryRowContext(ctx, `SELECT password_encrypted FROM birthday_settings`).Scan(&enc)
	if got, err := r.Settings(ctx); err != nil || got != s || enc != "enc:pw" {
		t.Fatalf("%+v %v %s", got, err, enc)
	}
}

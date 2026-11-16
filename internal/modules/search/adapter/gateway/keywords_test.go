package gateway

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/search/domain"
	"trisse-home/internal/modules/search/usecase"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestKeywordsPerAccount: own keywords (#478) are stored per account; a
// member starts with the admin's and gets an own copy on the first change.
func TestKeywordsPerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "search", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	kw := func(word string, kind domain.TargetKind, target string) domain.Keyword {
		k, err := domain.NewKeyword(word, kind, target, "Label "+word)
		if err != nil {
			t.Fatal(err)
		}
		return k
	}
	words := func(ctx context.Context) []string {
		list, err := repo.Keywords(ctx)
		if err != nil {
			t.Fatal(err)
		}
		out := []string{}
		for _, k := range list {
			out = append(out, k.Word()+"→"+string(k.Kind())+":"+k.Target())
		}
		return out
	}

	if err := repo.AddKeyword(admin, kw("nas", domain.TargetService, "svc-1")); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddKeyword(admin, kw("nas", domain.TargetURL, "https://x.example")); !errors.Is(err, usecase.ErrKeywordExists) {
		t.Fatalf("duplicate: %v", err)
	}
	if got := words(anna); len(got) != 1 || got[0] != "nas→service:svc-1" {
		t.Fatalf("member does not start with the admin's keywords: %v", got)
	}

	if err := repo.AddKeyword(anna, kw("bank", domain.TargetURL, "https://bank.example")); err != nil {
		t.Fatal(err)
	}
	if got := words(anna); len(got) != 2 {
		t.Fatalf("member's own list: %v", got)
	}
	if err := repo.DeleteKeyword(admin, "nas"); err != nil {
		t.Fatal(err)
	}
	if got := words(anna); len(got) != 2 {
		t.Fatalf("admin's change leaked into the member's list: %v", got)
	}
	if err := repo.DeleteKeyword(admin, "nas"); !errors.Is(err, usecase.ErrKeywordNotFound) {
		t.Fatalf("delete unknown: %v", err)
	}

	if err := repo.ReplaceKeywords(anna, []domain.Keyword{kw("wiki", domain.TargetQuery, "wiki ")}); err != nil {
		t.Fatal(err)
	}
	if got := words(anna); len(got) != 1 || got[0] != "wiki→query:wiki" {
		t.Fatalf("replace: %v", got)
	}
}

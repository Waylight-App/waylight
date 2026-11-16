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

// TestKeywordsArePersonal: own keywords (#478) can name private bookmarks,
// so unlike the shortcuts they are never inherited from the admin, guests
// and kiosk displays see none, and a deleted account's are removed.
func TestKeywordsArePersonal(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "search", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
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

	if err := repo.AddKeyword(admin, kw("bank", domain.TargetBookmark, "b-private")); err != nil {
		t.Fatal(err)
	}
	if err := repo.AddKeyword(admin, kw("bank", domain.TargetURL, "https://x.example")); !errors.Is(err, usecase.ErrKeywordExists) {
		t.Fatalf("duplicate: %v", err)
	}
	if got := words(anna); len(got) != 0 {
		t.Fatalf("a member must not see the owner's keywords: %v", got)
	}
	if got := words(guest); len(got) != 0 {
		t.Fatalf("a guest must not see the owner's keywords: %v", got)
	}
	if err := repo.AddKeyword(guest, kw("x", domain.TargetQuery, "x")); !errors.Is(err, kernel.ErrForbidden) {
		t.Fatalf("guest add: %v", err)
	}

	// A member's first change does not copy the owner's keywords either.
	if err := repo.AddKeyword(anna, kw("nas", domain.TargetURL, "https://nas.lan")); err != nil {
		t.Fatal(err)
	}
	if got := words(anna); len(got) != 1 || got[0] != "nas→url:https://nas.lan" {
		t.Fatalf("member's own list: %v", got)
	}
	if got := words(admin); len(got) != 1 || got[0] != "bank→bookmark:b-private" {
		t.Fatalf("owner's list changed: %v", got)
	}

	if err := repo.DeleteKeyword(anna, "bank"); !errors.Is(err, usecase.ErrKeywordNotFound) {
		t.Fatalf("a member deleted the owner's keyword: %v", err)
	}
	if err := repo.ReplaceKeywords(admin, []domain.Keyword{kw("wiki", domain.TargetQuery, "wiki ")}); err != nil {
		t.Fatal(err)
	}
	if got := words(admin); len(got) != 1 || got[0] != "wiki→query:wiki" {
		t.Fatalf("replace: %v", got)
	}

	if err := repo.ForgetAccount(context.Background(), "anna"); err != nil {
		t.Fatal(err)
	}
	if got := words(anna); len(got) != 0 {
		t.Fatalf("deleted account keeps keywords: %v", got)
	}
	if got := words(admin); len(got) != 1 {
		t.Fatalf("forgetting a member touched the owner's keywords: %v", got)
	}
}

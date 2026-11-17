package gateway

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/study/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestStudyIsPersonal: a transcript counts only the account's own modules,
// even when the admin's are visible to everyone (ADR 0018).
func TestStudyIsPersonal(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "study", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})

	if err := repo.Save(admin, domain.Module{ID: "m-admin", Name: "Maths", ECTS: 5, Status: domain.StatusOpen}); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(anna, domain.Module{ID: "m-anna", Name: "Art", ECTS: 5, Status: domain.StatusOpen}); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		ctx  context.Context
		want string
	}{"admin": {admin, "m-admin"}, "anna": {anna, "m-anna"}, "guest": {guest, "m-admin"}} {
		if list, err := repo.List(c.ctx); err != nil || len(list) != 1 || list[0].ID != c.want {
			t.Errorf("%s: %v %v", name, list, err)
		}
	}
	if err := repo.Save(anna, domain.Module{ID: "m-admin", Name: "Mine", ECTS: 1, Status: domain.StatusOpen}); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("anna changed the admin's module: %v", err)
	}
	if err := repo.Delete(anna, "m-admin"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("anna deleted the admin's module: %v", err)
	}
}

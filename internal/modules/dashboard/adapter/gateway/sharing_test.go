package gateway

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestTeamTabs: a tab the admin shares with a group is visible to its
// members only while the hook allows shared tabs, and never writable for
// them (#295).
func TestTeamTabs(t *testing.T) {
	on := false
	repo := NewRepository(dbtest.Open(t, "dashboard", Migrations())).WithSharedTabs(func(context.Context) bool { return on })
	admin := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner"})
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member", Groups: []string{"office"}})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})

	tools, _ := domain.NewTab("t-tools", "Office tools", 0)
	if err := repo.CreateTab(admin, tools); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetTabVisibility(anna, "t-tools", kernel.GroupVisibility("office")); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatal("only the owner shares a tab", err)
	}
	if err := repo.SetTabVisibility(admin, "t-tools", kernel.GroupVisibility("office")); err != nil {
		t.Fatal(err)
	}
	if own, _ := repo.OwnTabs(admin); len(own) != 1 || own[0].Access().Visibility != kernel.GroupVisibility("office") {
		t.Fatalf("own tabs: %+v", own)
	}

	if tabs, _ := repo.ListTabs(anna); len(tabs) != 0 {
		t.Fatalf("shared tabs are off without the hook: %+v", tabs)
	}
	on = true
	tabs, _ := repo.ListTabs(anna)
	if len(tabs) != 1 || tabs[0].Access().Mine(kernel.PrincipalFrom(anna)) {
		t.Fatalf("anna sees the team tab read-only: %+v", tabs)
	}
	if tabs, _ := repo.ListTabs(ben); len(tabs) != 0 {
		t.Fatalf("ben is not in the group: %+v", tabs)
	}
	renamed, _ := domain.NewTab("t-tools", "Mine now", 0)
	if err := repo.UpdateTab(anna, renamed); err == nil {
		if got, _ := repo.GetTab(admin, "t-tools"); got.Name() != "Office tools" {
			t.Fatal("a member changed the team tab")
		}
	}
}

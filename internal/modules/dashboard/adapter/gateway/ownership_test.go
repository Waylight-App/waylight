package gateway

import (
	"context"
	"errors"
	"slices"
	"testing"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

// TestPersonalTabs proves the separation of ADR 0018 for the dashboard: an
// account sees and changes only its own tabs and their widgets; guests see
// the tabs shared with them; the system sees all.
func TestPersonalTabs(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "dashboard", Migrations()))
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest", Tabs: []string{"home"}})
	system := context.Background()
	rect, _ := domain.NewRect(0, 0, 4, 2)

	mk := func(ctx context.Context, tabID, widgetID string) {
		t.Helper()
		tab, _ := domain.NewTab(tabID, tabID, 0)
		if err := repo.CreateTab(ctx, tab); err != nil {
			t.Fatal(err)
		}
		w, _ := domain.NewWidget(widgetID, tabID, "clock", rect, "{}")
		if err := repo.CreateWidget(ctx, w); err != nil {
			t.Fatal(err)
		}
	}
	mk(admin, "home", "w-home")
	mk(anna, "anna-tab", "w-anna")
	mk(ben, "ben-tab", "w-ben")

	ids := func(ctx context.Context) []string {
		tabs, err := repo.ListTabs(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, tb := range tabs {
			out = append(out, tb.ID())
		}
		return out
	}
	for name, c := range map[string]struct {
		ctx  context.Context
		want int
		has  string
	}{
		"admin": {admin, 1, "home"}, "anna": {anna, 1, "anna-tab"}, "ben": {ben, 1, "ben-tab"},
		"guest": {guest, 1, "home"}, "system": {system, 3, "ben-tab"},
	} {
		got := ids(c.ctx)
		if len(got) != c.want || !slices.Contains(got, c.has) {
			t.Errorf("%s sees %v", name, got)
		}
		ws, _ := repo.ListWidgets(c.ctx)
		if len(ws) != c.want {
			t.Errorf("%s sees %d widgets", name, len(ws))
		}
	}

	// Ben cannot read, change or delete Anna's tab or widget.
	if _, err := repo.GetTab(ben, "anna-tab"); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("ben read anna's tab: %v", err)
	}
	if _, err := repo.GetWidget(ben, "w-anna"); !errors.Is(err, domain.ErrWidgetNotFound) {
		t.Fatalf("ben read anna's widget: %v", err)
	}
	if ws, _ := repo.ListWidgetsByTab(ben, "anna-tab"); len(ws) != 0 {
		t.Fatalf("ben listed anna's widgets: %v", ws)
	}
	renamed := domain.RestoreTab("anna-tab", "mine now", 0, "")
	if err := repo.UpdateTab(ben, renamed); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("ben renamed anna's tab: %v", err)
	}
	intruder, _ := domain.NewWidget("w-x", "anna-tab", "clock", rect, "{}")
	if err := repo.CreateWidget(ben, intruder); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("ben added a widget to anna's tab: %v", err)
	}
	w, _ := repo.GetWidget(anna, "w-anna")
	big, _ := domain.NewRect(0, 0, 12, 9)
	w.Place(big)
	_ = repo.UpdateWidget(ben, w)
	_ = repo.SaveWidgetRects(ben, []domain.Widget{w})
	_ = repo.DeleteWidget(ben, "w-anna")
	_ = repo.DeleteTab(ben, "anna-tab")
	if got, err := repo.GetWidget(anna, "w-anna"); err != nil || got.Rect() != rect {
		t.Fatalf("anna's widget was changed by ben: %+v %v", got, err)
	}
	if got, err := repo.GetTab(anna, "anna-tab"); err != nil || got.Name() != "anna-tab" {
		t.Fatalf("anna's tab was changed by ben: %+v %v", got, err)
	}

	// A guest sees the shared tab but can change nothing.
	if _, err := repo.GetTab(guest, "anna-tab"); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatal("guest saw a tab not shared with it")
	}
	if err := repo.DeleteTab(guest, "home"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.GetTab(admin, "home"); err != nil {
		t.Fatal("a guest deleted the admin's tab")
	}

	// Config as code replaces the admin's tabs only.
	if err := repo.ReplaceAll(admin, []domain.Tab{domain.RestoreTab("new", "New", 0, "")}, nil); err != nil {
		t.Fatal(err)
	}
	if got := ids(admin); len(got) != 1 || got[0] != "new" {
		t.Fatalf("admin after replace: %v", got)
	}
	if got := ids(system); len(got) != 3 {
		t.Fatalf("members' tabs must survive the replace: %v", got)
	}
	if err := repo.ReplaceAll(admin, []domain.Tab{domain.RestoreTab("anna-tab", "Taken", 0, "")}, nil); !errors.Is(err, domain.ErrTabNotFound) {
		t.Fatalf("replace took over a member's tab: %v", err)
	}
}

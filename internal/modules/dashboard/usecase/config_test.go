package usecase

import (
	"context"
	"errors"
	"testing"

	"waylight/internal/modules/dashboard/domain"
)

type configRepo struct{ *memRepo }

func (c configRepo) ReplaceAll(_ context.Context, tabs []domain.Tab, widgets []domain.Widget) error {
	c.tabs = tabs
	c.widgets = map[string]domain.Widget{}
	for _, w := range widgets {
		c.widgets[w.ID()] = w
	}
	return nil
}

func TestDashboardConfig(t *testing.T) {
	repo := newMemRepo()
	old, _ := domain.NewTab("keep-me", "Home", 0)
	repo.tabs = []domain.Tab{old}
	uc := NewConfig(configRepo{repo}, kinds{"clock", "notes"}, &seqIDs{})
	in := []ConfigTab{{
		Name:  "home",
		Rules: []domain.RuleInput{{Days: []int{1, 2}, From: "07:00", To: "09:00"}},
		Widgets: []ConfigWidget{
			{Kind: "clock", Rect: domain.Rect{X: 0, Y: 0, W: 4, H: 2}, Phone: &ConfigPhone{Order: 2, Rows: 2}},
			{Kind: "notes", Rect: domain.Rect{X: 4, Y: 0, W: 4, H: 3}, Phone: &ConfigPhone{Order: 1, Rows: 5, Hidden: true}, Config: `{"a":1}`},
		},
	}, {Name: "Server"}}
	if err := uc.Apply(ctx, in); err != nil {
		t.Fatal(err)
	}
	if repo.tabs[0].ID() != "keep-me" || len(repo.tabs) != 2 || len(repo.widgets) != 2 {
		t.Fatalf("tab IDs are kept by name: %+v", repo.tabs)
	}
	out, err := uc.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	w := out[0].Widgets
	if w[0].Kind != "clock" || w[0].Phone.Order != 2 || w[1].Phone.Order != 1 || !w[1].Phone.Hidden || w[1].Config != `{"a":1}` {
		t.Fatalf("%+v", w)
	}
	if out[0].Rules[0].From != "07:00" || out[1].Widgets != nil {
		t.Fatalf("%+v", out)
	}
	bad := []ConfigTab{{Name: "X", Widgets: []ConfigWidget{{Kind: "weather", Rect: domain.Rect{W: 1, H: 1}}}}}
	if err := uc.Apply(ctx, bad); !errors.Is(err, domain.ErrUnknownWidgetKind) || len(repo.tabs) != 2 {
		t.Fatalf("a disabled widget kind must be an error: %v", err)
	}
	if err := uc.Apply(ctx, []ConfigTab{{Name: "A"}, {Name: "a"}}); err == nil {
		t.Fatal("duplicate tab names accepted")
	}
}

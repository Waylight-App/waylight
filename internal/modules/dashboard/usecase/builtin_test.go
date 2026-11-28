package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"waylight/internal/modules/dashboard/domain"
	"waylight/internal/shared/kernel"
)

func mergeForTest(config string, values map[string]any) (string, error) {
	// Enough for the test: the use case only hands the values on.
	if v, ok := values["label"].(string); ok {
		return strings.TrimSuffix(config, "}") + `,"label":"` + v + `"}`, nil
	}
	return config, nil
}

var familyTemplate = BuiltinTemplate{ID: "family", Title: "Family", Needs: []string{"location"}, Tab: ConfigTab{
	Name:  "Family",
	Zones: true,
	Widgets: []ConfigWidget{
		{Kind: "calendar", Rect: domain.Rect{X: 0, Y: 0, W: 6, H: 1}, Config: `{"view":"agenda"}`, Size: "strip", Zone: "today", Phone: &ConfigPhone{Order: 2, Rows: 5}},
		{Kind: "weather", Rect: domain.Rect{X: 6, Y: 0, W: 3, H: 2}, Config: `{"days":3}`, Size: "small", Zone: "mine", Phone: &ConfigPhone{Order: 1, Rows: 3}},
		{Kind: "waste", Rect: domain.Rect{X: 9, Y: 0, W: 3, H: 2}, Config: `{}`, Phone: &ConfigPhone{Order: 3, Rows: 2}},
	},
}}

// TestBuiltinCreateTab: a template becomes a tab with its widgets and phone
// layout, zones and sizes (#670); widgets of disabled modules are left out
// and the weather widget gets the location (#589).
func TestBuiltinCreateTab(t *testing.T) {
	deps, repo := newDeps("calendar", "weather")
	uc := NewBuiltinTemplates(deps, []BuiltinTemplate{familyTemplate}, mergeForTest, nil)
	if len(uc.List()) != 1 {
		t.Fatal("list")
	}
	res, err := uc.CreateTab(ctx, "family", "", &Location{Lat: 53.1, Lon: 8.2, Label: "Home"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Name != "Family" || len(res.Widgets) != 2 {
		t.Fatalf("%+v", res)
	}
	for _, w := range res.Widgets {
		if w.Kind == "weather" && !strings.Contains(w.Config, `"label":"Home"`) {
			t.Fatalf("weather config %s", w.Config)
		}
		if w.Kind == "calendar" && strings.Contains(w.Config, "label") {
			t.Fatalf("calendar config %s", w.Config)
		}
	}
	tab, _ := repo.GetTab(ctx, res.ID)
	if !tab.Zoned() || tab.Sample() {
		t.Fatalf("zoned %v, sample %v", tab.Zoned(), tab.Sample())
	}
	for _, w := range repo.widgets {
		if w.Kind() == "weather" && (w.Size() != domain.SizeSmall || w.Zone() != domain.ZoneMine) {
			t.Fatalf("weather size %q zone %q", w.Size(), w.Zone())
		}
	}
	m := tab.Mobile()
	if m == nil || len(m) != 2 {
		t.Fatalf("phone layout %+v", m)
	}
	// The weather widget comes first on phones, as the template says.
	first := repo.widgets[m[0].WidgetID]
	if first.Kind() != "weather" {
		t.Fatalf("phone order starts with %s", first.Kind())
	}
	if _, err := uc.CreateTab(ctx, "nope", "", nil, ""); !errors.Is(err, ErrBuiltinNotFound) {
		t.Fatalf("unknown: %v", err)
	}
	if _, err := uc.CreateTab(ctx, "family", "", &Location{Lat: 95}, ""); err == nil {
		t.Fatal("invalid location")
	}
}

// TestBuiltinShare: a template that needs "share" sets up the shared
// records of its widgets with the chosen visibility and merges what the
// setup returns into the widget (#871). The visibility is checked like any
// other share, and templates without the need share nothing.
func TestBuiltinShare(t *testing.T) {
	deps, _ := newDeps("calendar", "weather")
	shared := familyTemplate
	shared.Needs = []string{"location", "share"}
	var calls []string
	setup := func(_ context.Context, kind, tabName string, v kernel.Visibility) (map[string]any, error) {
		calls = append(calls, kind+" "+tabName+" "+string(v))
		if kind == "weather" {
			return map[string]any{"label": "Shared"}, nil
		}
		return nil, nil
	}
	uc := NewBuiltinTemplates(deps, []BuiltinTemplate{shared}, mergeForTest, setup)
	anna := kernel.WithPrincipal(ctx, kernel.Principal{AccountID: "anna", Role: "member", Groups: []string{"family"}})

	res, err := uc.CreateTab(anna, "family", "Home", nil, "group:family")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(calls, ",") != "calendar Home group:family,weather Home group:family" {
		t.Fatalf("setup calls %v", calls)
	}
	for _, w := range res.Widgets {
		if w.Kind == "weather" && !strings.Contains(w.Config, `"label":"Shared"`) {
			t.Fatalf("weather config %s", w.Config)
		}
	}

	calls = nil
	if _, err := uc.CreateTab(anna, "family", "", nil, "group:office"); !errors.Is(err, kernel.ErrShareNotAllowed) {
		t.Fatalf("foreign group: %v", err)
	}
	if len(calls) != 0 {
		t.Fatalf("set up despite refused share: %v", calls)
	}
	if _, err := uc.CreateTab(anna, "family", "", nil, ""); err != nil || len(calls) != 0 {
		t.Fatalf("no share: %v %v", err, calls)
	}

	plain := NewBuiltinTemplates(deps, []BuiltinTemplate{familyTemplate}, mergeForTest, setup)
	if _, err := plain.CreateTab(anna, "family", "", nil, "everyone"); err != nil || len(calls) != 0 {
		t.Fatalf("template without share: %v %v", err, calls)
	}
}

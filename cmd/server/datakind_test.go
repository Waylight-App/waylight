package main

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"testing"

	"trisse-home/internal/modules"
)

// TestDataKinds keeps the kinds the modules declare in step with the table
// in ADR 0018. Moving a module to another kind changes who may read its
// data, so it has to be a visible change here and in the ADR.
func TestDataKinds(t *testing.T) {
	up := fakeUpstream(t)
	a, err := build(context.Background(), newTestConfig(t, up.URL), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer a.db.Close()
	want := map[modules.DataKind][]string{
		modules.PersonalContent: {"birthdays", "bookmarks", "countdown", "dashboard", "drop", "focus", "gamedeals", "habits",
			"markets", "notes", "parcels", "photos", "readlater", "releases", "rss", "shopping", "study", "todos"},
		modules.PersonalSettings: {"appearance", "briefing", "calendar", "mail", "onboarding", "search"},
	}
	got := map[modules.DataKind][]string{}
	for _, m := range a.registry.All() {
		switch k := m.DataKind(); k {
		case modules.InstanceData, modules.PersonalSettings, modules.PersonalContent:
			got[k] = append(got[k], m.Name())
		default:
			t.Errorf("%s: unknown data kind %q", m.Name(), k)
		}
	}
	for _, k := range []modules.DataKind{modules.PersonalContent, modules.PersonalSettings} {
		slices.Sort(got[k])
		if !slices.Equal(got[k], want[k]) {
			t.Errorf("%s:\n got  %v\n want %v (ADR 0018)", k, got[k], want[k])
		}
	}
}

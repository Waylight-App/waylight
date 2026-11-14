package domain

import (
	"strings"
	"testing"
)

func TestMergeTilesOnlyAdds(t *testing.T) {
	cur := []any{map[string]any{"name": "NAS", "url": "https://nas"}}
	out, added, skipped := MergeTiles(cur, []ImportTile{{Name: "nas", URL: "https://other"}, {Name: "Wiki", URL: "https://wiki", Icon: "https://wiki/i.png"}})
	if added != 1 || skipped != 1 || len(out) != 2 || out[1].(map[string]any)["icon"] != "https://wiki/i.png" {
		t.Fatalf("%v %d %d", out, added, skipped)
	}
}

func TestMergeFolders(t *testing.T) {
	cur := []any{
		map[string]any{"url": "https://loose"},
		map[string]any{"folder": "Dev", "bookmarks": []any{map[string]any{"url": "https://github.com/"}}},
	}
	out, added, skipped := MergeFolders(cur, []ImportFolder{
		{Name: "dev", Links: []ImportLink{{Title: "GH", URL: "https://github.com/"}, {Title: "Go", URL: "https://go.dev"}}},
		{Name: "New", Links: []ImportLink{{URL: "https://loose"}, {URL: "https://new"}}},
	})
	if added != 2 || skipped != 2 || len(out) != 3 {
		t.Fatalf("%v %d %d", out, added, skipped)
	}
	if dev := out[1].(map[string]any)["bookmarks"].([]any); len(dev) != 2 {
		t.Fatalf("existing folder extended: %v", dev)
	}
}

func TestAddImportTab(t *testing.T) {
	dash := map[string]any{"tabs": []any{map[string]any{"name": "Home"}}}
	plan := ImportPlan{Source: "Homepage", Widgets: []ImportWidget{{Kind: "clock"}, {Kind: "weather", Config: map[string]any{"lat": 1.0}}, {Kind: "system"}}}
	available := func(k string) bool { return k != "weather" }
	name, n, warnings := AddImportTab(dash, plan, true, true, available)
	if name != "Homepage" || n != 4 || len(warnings) != 1 || !strings.Contains(warnings[0], "weather") {
		t.Fatalf("%q %d %v", name, n, warnings)
	}
	widgets := dash["tabs"].([]any)[1].(map[string]any)["widgets"].([]any)
	at := func(i int) map[string]any { return widgets[i].(map[string]any)["at"].(map[string]any) }
	// services 8 wide and bookmarks 4 wide fill the first row; clock and system go below.
	if at(1)["x"] != 8 || at(2)["y"] != 4 || at(3)["x"] != 4 {
		t.Fatalf("layout: %v", widgets)
	}
	if name, n, w := AddImportTab(dash, plan, true, true, available); name != "" || n != 0 || len(w) != 1 {
		t.Fatalf("a second import must not add the tab again: %q %d %v", name, n, w)
	}
	if name, n, _ := AddImportTab(map[string]any{}, ImportPlan{Source: "X"}, false, false, available); name != "" || n != 0 {
		t.Fatal("nothing to place means no tab")
	}
}

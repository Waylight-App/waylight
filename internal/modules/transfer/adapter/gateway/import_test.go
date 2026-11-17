package gateway

import (
	"errors"
	"os"
	"strings"
	"testing"

	"waylight/internal/modules/transfer/domain"
	"waylight/internal/shared/kernel"
)

// The fixtures follow the examples of the Homepage documentation and the
// board file layout of Homarr 0.x (data/configs/default.json).

func read(t *testing.T, names ...string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for _, n := range names {
		b, err := os.ReadFile("testdata/" + n)
		if err != nil {
			t.Fatal(err)
		}
		out[n] = b
	}
	return out
}

func hasWarning(p domain.ImportPlan, part string) bool {
	for _, w := range p.Warnings {
		if strings.Contains(w, part) {
			return true
		}
	}
	return false
}

func TestHomepage(t *testing.T) {
	p, err := Homepage{}.Parse(read(t, "homepage/services.yaml", "homepage/bookmarks.yaml", "homepage/widgets.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tile := range p.Tiles {
		names = append(names, tile.Name)
	}
	// Nested groups are flattened in order; entries without link are skipped.
	if strings.Join(names, ",") != "Jellyfin,Sonarr,Router,Pi-hole" {
		t.Fatalf("tiles: %v", names)
	}
	if p.Tiles[0].Icon != "" || p.Tiles[1].Icon == "" || p.Tiles[0].Description != "Movies and series" {
		t.Fatalf("icons by name are not taken, URLs are: %+v", p.Tiles[:2])
	}
	if len(p.Folders) != 2 || p.Folders[0].Name != "Developer" || p.Folders[1].Links[0].URL != "https://reddit.com/" || len(p.Folders[1].Links) != 1 {
		t.Fatalf("folders: %+v", p.Folders)
	}
	var kinds []string
	for _, w := range p.Widgets {
		kinds = append(kinds, w.Kind)
	}
	if strings.Join(kinds, ",") != "system,weather,clock" || p.Widgets[1].Config["lat"] != 52.52 {
		t.Fatalf("widgets: %+v", p.Widgets)
	}
	for _, want := range []string{"jellyfin widget of Jellyfin", "Broken skipped", "Nothing skipped", "search skipped", "greeting skipped", "jellyfin.png"} {
		if !hasWarning(p, want) {
			t.Errorf("missing warning %q in %v", want, p.Warnings)
		}
	}
}

func TestHomepageRejects(t *testing.T) {
	if _, err := (Homepage{}).Parse(map[string][]byte{"settings.yaml": []byte("title: x")}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("no known file: %v", err)
	}
	if _, err := (Homepage{}).Parse(map[string][]byte{"services.yaml": []byte("- a: [")}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("broken YAML: %v", err)
	}
}

func TestHomarr(t *testing.T) {
	p, err := Homarr{}.Parse(read(t, "homarr/default.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Tiles) != 2 || p.Tiles[0].URL != "https://plex.example.org" || p.Tiles[1].URL != "http://192.168.1.20:8080" || p.Tiles[0].Icon == "" || p.Tiles[1].Icon != "" {
		t.Fatalf("tiles: %+v", p.Tiles)
	}
	if len(p.Widgets) != 2 || p.Widgets[0].Kind != "clock" || p.Widgets[1].Config["label"] != "Paris" {
		t.Fatalf("widgets: %+v", p.Widgets)
	}
	for _, want := range []string{"No link skipped", "torrents-status skipped", "Categories"} {
		if !hasWarning(p, want) {
			t.Errorf("missing warning %q in %v", want, p.Warnings)
		}
	}
	if _, err := (Homarr{}).Parse(map[string][]byte{"x.json": []byte(`{"board":{}}`)}); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("a Homarr 1.x or foreign file must be refused: %v", err)
	}
}

package domain

import (
	"fmt"
	"strings"
)

// ImportTile is a service tile found in another dashboard (#77).
type ImportTile struct {
	Name        string
	URL         string
	Icon        string
	Description string
}

// ImportLink is a bookmark found in another dashboard.
type ImportLink struct {
	Title string
	URL   string
}

// ImportFolder groups imported bookmarks.
type ImportFolder struct {
	Name  string
	Links []ImportLink
}

// ImportWidget is an information widget that has a trisse-home counterpart.
type ImportWidget struct {
	Kind   string
	Config map[string]any
}

// ImportPlan is what a foreign format contains, independent of the format.
// Warnings name everything that could not be carried over.
type ImportPlan struct {
	// Source names the format, e.g. "Homepage"; it becomes the tab name.
	Source   string
	Tiles    []ImportTile
	Folders  []ImportFolder
	Widgets  []ImportWidget
	Warnings []string
}

// ImportSummary says what merging a plan added.
type ImportSummary struct {
	Tiles     int
	Bookmarks int
	Widgets   int
	// Skipped counts tiles and bookmarks that already existed.
	Skipped int
	// Tab is the name of the new tab ("" when none was added).
	Tab      string
	Warnings []string
}

// Imports only ever add: an existing tile, bookmark or tab is never
// changed, so importing twice does no harm. The values below are the
// generic form of the sections of trisse-home.yaml (maps and slices).

func str(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}

// MergeTiles appends tiles whose name is not taken yet.
func MergeTiles(services []any, tiles []ImportTile) ([]any, int, int) {
	have := map[string]bool{}
	for _, s := range services {
		if m, ok := s.(map[string]any); ok {
			have[strings.ToLower(str(m, "name"))] = true
		}
	}
	added, skipped := 0, 0
	for _, t := range tiles {
		key := strings.ToLower(strings.TrimSpace(t.Name))
		if key == "" || have[key] {
			skipped++
			continue
		}
		have[key] = true
		m := map[string]any{"name": strings.TrimSpace(t.Name), "url": t.URL}
		if t.Icon != "" {
			m["icon"] = t.Icon
		}
		if t.Description != "" {
			m["description"] = t.Description
		}
		services = append(services, m)
		added++
	}
	return services, added, skipped
}

// MergeFolders adds bookmarks to folders of the same name, or new folders,
// skipping URLs that are already bookmarked anywhere.
func MergeFolders(bookmarks []any, folders []ImportFolder) ([]any, int, int) {
	urls := map[string]bool{}
	folderAt := map[string]int{}
	for i, e := range bookmarks {
		m, ok := e.(map[string]any)
		if !ok {
			continue
		}
		if u := str(m, "url"); u != "" {
			urls[u] = true
		}
		if f := str(m, "folder"); f != "" {
			folderAt[strings.ToLower(f)] = i
			list, _ := m["bookmarks"].([]any)
			for _, b := range list {
				if bm, ok := b.(map[string]any); ok {
					urls[str(bm, "url")] = true
				}
			}
		}
	}
	added, skipped := 0, 0
	for _, f := range folders {
		var links []any
		for _, l := range f.Links {
			if l.URL == "" || urls[l.URL] {
				skipped++
				continue
			}
			urls[l.URL] = true
			b := map[string]any{"url": l.URL}
			if l.Title != "" {
				b["title"] = l.Title
			}
			links = append(links, b)
			added++
		}
		if len(links) == 0 {
			continue
		}
		if i, ok := folderAt[strings.ToLower(f.Name)]; ok {
			m := bookmarks[i].(map[string]any)
			existing, _ := m["bookmarks"].([]any)
			m["bookmarks"] = append(existing, links...)
			continue
		}
		folderAt[strings.ToLower(f.Name)] = len(bookmarks)
		bookmarks = append(bookmarks, map[string]any{"folder": f.Name, "bookmarks": links})
	}
	return bookmarks, added, skipped
}

// Widget sizes on the import tab, in grid cells of the 12-column grid.
const (
	importWide   = 8
	importNarrow = 4
	importTall   = 4
	importShort  = 3
	gridColumns  = 12
)

// AddImportTab appends a tab named after the source with the services and
// bookmarks widgets (when tiles or bookmarks were imported) and the mapped
// information widgets. Widgets of disabled modules become warnings, since
// the dashboard refuses kinds it does not know. If the tab exists already
// (an earlier import), nothing is added, so importing twice does no harm.
func AddImportTab(dashboard map[string]any, plan ImportPlan, withTiles, withBookmarks bool, available func(string) bool) (string, int, []string) {
	tabs, _ := dashboard["tabs"].([]any)
	for _, t := range tabs {
		if m, ok := t.(map[string]any); ok && strings.EqualFold(str(m, "name"), plan.Source) {
			return "", 0, []string{fmt.Sprintf("The tab %s exists already; no widgets were added.", str(m, "name"))}
		}
	}
	var warnings []string
	var widgets []ImportWidget
	if withTiles {
		widgets = append(widgets, ImportWidget{Kind: "services"})
	}
	if withBookmarks {
		widgets = append(widgets, ImportWidget{Kind: "bookmarks"})
	}
	widgets = append(widgets, plan.Widgets...)
	var placed []any
	x, y, rowH := 0, 0, 0
	for _, w := range widgets {
		if !available(w.Kind) {
			warnings = append(warnings, fmt.Sprintf("Widget %s skipped: its module is disabled.", w.Kind))
			continue
		}
		width, height := importNarrow, importShort
		switch w.Kind {
		case "services":
			width, height = importWide, importTall
		case "bookmarks":
			height = importTall
		}
		if x+width > gridColumns {
			x, y, rowH = 0, y+rowH, 0
		}
		m := map[string]any{"kind": w.Kind, "at": map[string]any{"x": x, "y": y, "w": width, "h": height}}
		if len(w.Config) > 0 {
			m["config"] = w.Config
		}
		placed = append(placed, m)
		x += width
		rowH = max(rowH, height)
	}
	if len(placed) == 0 {
		return "", 0, warnings
	}
	dashboard["tabs"] = append(tabs, map[string]any{"name": plan.Source, "widgets": placed})
	return plan.Source, len(placed), warnings
}

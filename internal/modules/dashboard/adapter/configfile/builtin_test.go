package configfile

import (
	"cmp"
	"slices"
	"testing"

	"waylight/internal/modules/dashboard/usecase"
)

// webCells are the grid cells of the web sizes (#595), as the web client
// sizes a tile; a sized widget of a template must have exactly these.
var webCells = map[string][2]int{"small": {3, 2}, "medium": {4, 4}, "large": {6, 6}, "strip": {6, 1}}

// TestBuiltin checks the shipped templates (#589): each file parses with
// known fields only, has an ID and a title, places its widgets on the grid
// without overlap and gives every widget a phone slot. Each template shows
// the zoned layout (#619, #670): zones on, every widget in a zone, a status
// widget in the System zone, and sized widgets with the cells of their size.
func TestBuiltin(t *testing.T) {
	list, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, b := range list {
		ids = append(ids, b.ID)
		if b.ID == "" || b.Title == "" || b.Tab.Name == "" || len(b.Tab.Widgets) == 0 {
			t.Errorf("%s: incomplete %+v", b.ID, b)
		}
		if !b.Tab.Zones {
			t.Errorf("%s: not zoned", b.ID)
		}
		system := false
		for i, w := range b.Tab.Widgets {
			r := w.Rect
			if w.Zone == "" {
				t.Errorf("%s: %s has no zone", b.ID, w.Kind)
			}
			system = system || w.Zone == "system"
			if cells, ok := webCells[w.Size]; ok && (r.W != cells[0] || r.H != cells[1]) {
				t.Errorf("%s: %s is %s but %dx%d", b.ID, w.Kind, w.Size, r.W, r.H)
			}
			if r.X < 0 || r.W < 1 || r.X+r.W > 12 || r.H < 1 {
				t.Errorf("%s: %s outside the grid: %+v", b.ID, w.Kind, r)
			}
			if w.Phone == nil {
				t.Errorf("%s: %s has no phone slot", b.ID, w.Kind)
			}
			for _, o := range b.Tab.Widgets[i+1:] {
				q := o.Rect
				if r.X < q.X+q.W && q.X < r.X+r.W && r.Y < q.Y+q.H && q.Y < r.Y+r.H {
					t.Errorf("%s: %s overlaps %s", b.ID, w.Kind, o.Kind)
				}
			}
		}
		if !system {
			t.Errorf("%s: nothing in the System zone", b.ID)
		}
	}
	if !slices.Equal(ids, []string{"family", "homelab", "wall", "work"}) {
		t.Fatalf("templates: %v", ids)
	}
}

// TestBuiltinRowsFill checks that every row of every zone of a template
// adds up to the 12 columns (#793): the widgets of a zone cover each cell of
// the band between its top and bottom row exactly once. Zones lay out their
// widgets in reading order with a dense 12-column grid, so the band must
// also come out of that flow unchanged, or the zoned view would leave gaps
// the free grid does not have.
func TestBuiltinRowsFill(t *testing.T) {
	list, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range list {
		zones := map[string][]usecase.ConfigWidget{}
		for _, w := range b.Tab.Widgets {
			zones[w.Zone] = append(zones[w.Zone], w)
		}
		for zone, ws := range zones {
			slices.SortStableFunc(ws, func(a, c usecase.ConfigWidget) int {
				return cmp.Or(cmp.Compare(a.Rect.Y, c.Rect.Y), cmp.Compare(a.Rect.X, c.Rect.X))
			})
			top, bottom := ws[0].Rect.Y, 0
			for _, w := range ws {
				bottom = max(bottom, w.Rect.Y+w.Rect.H)
			}
			cells := make([][12]int, bottom-top)
			for _, w := range ws {
				r := w.Rect
				for y := r.Y; y < r.Y+r.H; y++ {
					for x := r.X; x < r.X+r.W && x < 12; x++ {
						cells[y-top][x]++
					}
				}
			}
			for y, row := range cells {
				for x, n := range row {
					if n != 1 {
						t.Errorf("%s/%s: cell %d,%d covered %d times", b.ID, zone, x, y+top, n)
					}
				}
			}
			for i, at := range denseFlow(ws) {
				if r := ws[i].Rect; at != [2]int{r.X, r.Y - top} {
					t.Errorf("%s/%s: %s flows to %v, not %d,%d", b.ID, zone, ws[i].Kind, at, r.X, r.Y-top)
				}
			}
		}
	}
}

// denseFlow places widgets the way CSS grid auto-flow "row dense" does on 12
// columns: each one at the first free spot from the top left.
func denseFlow(ws []usecase.ConfigWidget) [][2]int {
	var used [][12]bool
	fits := func(x, y, w, h int) bool {
		for yy := y; yy < y+h; yy++ {
			for xx := x; xx < x+w; xx++ {
				if yy < len(used) && used[yy][xx] {
					return false
				}
			}
		}
		return true
	}
	out := make([][2]int, 0, len(ws))
	for _, wd := range ws {
		w, h := min(wd.Rect.W, 12), wd.Rect.H
		for y := 0; ; y++ {
			x, placed := 0, false
			for ; x+w <= 12; x++ {
				if fits(x, y, w, h) {
					placed = true
					break
				}
			}
			if !placed {
				continue
			}
			for len(used) < y+h {
				used = append(used, [12]bool{})
			}
			for yy := y; yy < y+h; yy++ {
				for xx := x; xx < x+w; xx++ {
					used[yy][xx] = true
				}
			}
			out = append(out, [2]int{x, y})
			break
		}
	}
	return out
}

// TestFamilyWidgets checks that the family template gives the household
// widgets of #675 and #676 a place.
func TestFamilyWidgets(t *testing.T) {
	list, err := Builtin()
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range list {
		if b.ID != "family" {
			continue
		}
		var kinds []string
		for _, w := range b.Tab.Widgets {
			kinds = append(kinds, w.Kind)
		}
		for _, want := range []string{"meals", "chores", "board", "parcels", "waste"} {
			if !slices.Contains(kinds, want) {
				t.Errorf("family has no %s widget", want)
			}
		}
		return
	}
	t.Fatal("no family template")
}

func TestMergeConfig(t *testing.T) {
	got, err := MergeConfig(`{"days":4,"lat":1}`, map[string]any{"lat": 53.1, "label": "Home"})
	if err != nil || got != `{"days":4,"label":"Home","lat":53.1}` {
		t.Fatalf("%s %v", got, err)
	}
	if _, err := MergeConfig(`not json`, nil); err == nil {
		t.Fatal("broken settings")
	}
}

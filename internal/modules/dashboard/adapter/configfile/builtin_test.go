package configfile

import (
	"slices"
	"testing"
)

// TestBuiltin checks the shipped templates (#589): each file parses with
// known fields only, has an ID and a title, places its widgets on the grid
// without overlap and gives every widget a phone slot.
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
		for i, w := range b.Tab.Widgets {
			r := w.Rect
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
	}
	if !slices.Equal(ids, []string{"family", "homelab", "wall", "work"}) {
		t.Fatalf("templates: %v", ids)
	}
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

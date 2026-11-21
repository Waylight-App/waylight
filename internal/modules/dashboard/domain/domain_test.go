package domain

import (
	"errors"
	"strings"
	"testing"

	"waylight/internal/shared/kernel"
)

func TestNewRect(t *testing.T) {
	tests := []struct {
		name       string
		x, y, w, h int
		wantErr    bool
	}{
		{"full width", 0, 0, 12, 2, false},
		{"zero width", 0, 0, 0, 2, true},
		{"overflow right", 8, 0, 5, 2, true},
		{"negative x", -1, 0, 2, 2, true},
		{"too tall", 0, 0, 2, 25, true},
		{"row out of range", 0, 501, 2, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRect(tt.x, tt.y, tt.w, tt.h)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRectOverlaps(t *testing.T) {
	a := Rect{X: 0, Y: 0, W: 4, H: 2}
	if !a.Overlaps(Rect{X: 3, Y: 1, W: 2, H: 2}) {
		t.Fatal("expected overlap")
	}
	if a.Overlaps(Rect{X: 4, Y: 0, W: 2, H: 2}) || a.Overlaps(Rect{X: 0, Y: 2, W: 2, H: 2}) {
		t.Fatal("adjacent rectangles must not overlap")
	}
}

func TestTabName(t *testing.T) {
	if _, err := NewTab("1", "   ", 0); !errors.Is(err, kernel.ErrInvalid) {
		t.Fatalf("expected invalid, got %v", err)
	}
	if _, err := NewTab("1", strings.Repeat("x", 41), 0); err == nil {
		t.Fatal("expected error for long name")
	}
	tab, err := NewTab("1", "  Server ", 0)
	if err != nil || tab.Name() != "Server" {
		t.Fatalf("name not trimmed: %q %v", tab.Name(), err)
	}
}

func TestWidget(t *testing.T) {
	r, _ := NewRect(0, 0, 2, 2)
	if _, err := NewWidget("1", "t", "Bad Kind", r, ""); err == nil {
		t.Fatal("expected invalid kind")
	}
	w, err := NewWidget("1", "t", "weather", r, "")
	if err != nil || w.Config() != "{}" {
		t.Fatalf("empty config not normalised: %q %v", w.Config(), err)
	}
	if err := w.Configure(strings.Repeat("x", MaxConfigBytes+1)); err == nil {
		t.Fatal("expected size error")
	}
}

func TestWidgetSize(t *testing.T) {
	r, _ := NewRect(0, 0, 3, 2)
	w, _ := NewWidget("1", "t", "todos", r, "")
	if w.Size() != SizeAuto {
		t.Fatalf("new widget size = %q, want free", w.Size())
	}
	if err := w.Resize(SizeSmall); err != nil || w.Size() != SizeSmall {
		t.Fatalf("resize to small: %q %v", w.Size(), err)
	}
	if err := w.Resize("huge"); err == nil || w.Size() != SizeSmall {
		t.Fatalf("unknown size accepted: %q %v", w.Size(), err)
	}
	if got := RestoreWidget("1", "t", "todos", r, "{}", "huge").Size(); got != SizeAuto {
		t.Fatalf("restored unknown size = %q, want free", got)
	}
}

func TestWidgetZone(t *testing.T) {
	r, _ := NewRect(0, 0, 3, 2)
	w, _ := NewWidget("1", "t", "services", r, "")
	if w.Zone() != ZoneAuto {
		t.Fatalf("new widget zone = %q, want by kind", w.Zone())
	}
	if err := w.MoveToZone(ZoneSystem); err != nil || w.Zone() != ZoneSystem {
		t.Fatalf("move to system: %q %v", w.Zone(), err)
	}
	if err := w.MoveToZone("attic"); err == nil || w.Zone() != ZoneSystem {
		t.Fatalf("unknown zone accepted: %q %v", w.Zone(), err)
	}
	if got := w.WithZone("attic").Zone(); got != ZoneAuto {
		t.Fatalf("restored unknown zone = %q, want by kind", got)
	}
	if tab := RestoreTab("t", "Home", 0, ""); tab.Zoned() || !tab.WithZones(true).Zoned() {
		t.Fatal("zones switch")
	}
}

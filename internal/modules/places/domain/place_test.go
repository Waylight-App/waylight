package domain

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestNewPlace(t *testing.T) {
	p, err := NewPlace("1", "  Home ", " Oldenburg ", 53.14, 8.21, time.Unix(0, 0))
	if err != nil || p.Name != "Home" || p.Label != "Oldenburg" {
		t.Fatalf("%+v %v", p, err)
	}
	bad := []struct {
		name     string
		lat, lon float64
	}{
		{"", 0, 0}, {strings.Repeat("x", MaxNameLength+1), 0, 0},
		{"A", 91, 0}, {"A", 0, -181}, {"A", math.NaN(), 0},
	}
	for _, b := range bad {
		if _, err := NewPlace("1", b.name, "", b.lat, b.lon, time.Unix(0, 0)); err == nil {
			t.Errorf("%+v accepted", b)
		}
	}
	if _, err := NewPlace("1", "A", strings.Repeat("x", MaxLabelLength+1), 0, 0, time.Unix(0, 0)); err == nil {
		t.Error("long label accepted")
	}
}

func TestChangeKeepsThePlaceOnError(t *testing.T) {
	p, _ := NewPlace("1", "Home", "", 1, 2, time.Unix(0, 0))
	if err := p.Change("", "", 3, 4); err == nil || p.Name != "Home" || p.Lat != 1 {
		t.Fatalf("%+v %v", p, err)
	}
}

package domain

import (
	"math"
	"testing"
	"time"
)

func TestNewLocation(t *testing.T) {
	tests := []struct {
		lat, lon float64
		wantErr  bool
		wantKey  string
	}{
		{53.1435, 8.2146, false, "53.14,8.21"},
		{-91, 0, true, ""},
		{0, 181, true, ""},
		{math.NaN(), 0, true, ""},
	}
	for _, tt := range tests {
		loc, err := NewLocation(tt.lat, tt.lon)
		if (err != nil) != tt.wantErr || (err == nil && loc.Key() != tt.wantKey) {
			t.Errorf("NewLocation(%v,%v) = %v, %v", tt.lat, tt.lon, loc.Key(), err)
		}
	}
}

func TestNormalizeDegrees(t *testing.T) {
	for in, want := range map[float64]float64{0: 0, 90: 90, 359.5: 359.5, 360: 0, 450: 90, -45: 315, -360: 0, 720.5: 0.5} {
		if got := NormalizeDegrees(in); got != want {
			t.Errorf("NormalizeDegrees(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestIsDayAt(t *testing.T) {
	at := func(day, h, m int) time.Time { return time.Date(2026, 11, day, h, m, 0, 0, time.UTC) }
	f := Forecast{
		// The service said "day" at the fetch; the sun times decide later.
		Current: Current{IsDay: true},
		Days: []Day{
			{Sunrise: at(28, 7, 0), Sunset: at(28, 16, 0)},
			{Sunrise: at(29, 7, 2), Sunset: at(29, 15, 59)},
		},
	}
	for name, c := range map[string]struct {
		now  time.Time
		want bool
	}{
		"before the first sunrise": {at(28, 5, 0), false},
		"at sunrise":               {at(28, 7, 0), true},
		"afternoon":                {at(28, 15, 59), true},
		"at sunset":                {at(28, 16, 0), false},
		"night between two days":   {at(29, 2, 0), false},
		"second day":               {at(29, 12, 0), true},
		"past the forecast":        {at(30, 20, 0), true},
	} {
		if got := f.IsDayAt(c.now); got != c.want {
			t.Errorf("%s: IsDayAt = %v, want %v", name, got, c.want)
		}
	}
	if !(Forecast{Current: Current{IsDay: true}, Days: []Day{{}}}).IsDayAt(at(28, 23, 0)) {
		t.Error("without sun times the service's flag must be used")
	}
}

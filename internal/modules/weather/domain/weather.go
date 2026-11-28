package domain

import (
	"fmt"
	"math"
	"time"

	"waylight/internal/shared/kernel"
)

// coordinatePrecision rounds coordinates to ~1 km so nearby widgets share a
// cache entry and an external request.
const coordinatePrecision = 100

// Location is a validated pair of WGS84 coordinates.
type Location struct {
	lat, lon float64
}

// NewLocation validates the coordinates and rounds them to ~1 km.
func NewLocation(lat, lon float64) (Location, error) {
	if math.IsNaN(lat) || lat < -90 || lat > 90 {
		return Location{}, kernel.NewValidationError("lat", "The latitude must be between -90 and 90.")
	}
	if math.IsNaN(lon) || lon < -180 || lon > 180 {
		return Location{}, kernel.NewValidationError("lon", "The longitude must be between -180 and 180.")
	}
	round := func(v float64) float64 { return math.Round(v*coordinatePrecision) / coordinatePrecision }
	return Location{lat: round(lat), lon: round(lon)}, nil
}

// Lat returns the latitude.
func (l Location) Lat() float64 { return l.lat }

// Lon returns the longitude.
func (l Location) Lon() float64 { return l.lon }

// Key identifies the location in caches.
func (l Location) Key() string { return fmt.Sprintf("%.2f,%.2f", l.lat, l.lon) }

// Current describes the weather right now. WeatherCode follows the WMO code
// table used by Open-Meteo.
type Current struct {
	Temperature         float64
	ApparentTemperature float64
	Humidity            float64
	WindSpeed           float64
	// WindDirection is where the wind comes from in degrees (0 = north,
	// 90 = east), in [0, 360). It is only meaningful when HasWindDirection.
	WindDirection    float64
	HasWindDirection bool
	WeatherCode      int
	IsDay            bool
}

// NormalizeDegrees maps any angle into [0, 360).
func NormalizeDegrees(degrees float64) float64 {
	d := math.Mod(degrees, 360)
	if d < 0 {
		d += 360
	}
	return d
}

// Day is the forecast for one day.
type Day struct {
	Date                     time.Time
	TemperatureMin           float64
	TemperatureMax           float64
	WeatherCode              int
	PrecipitationProbability float64
	// Sunrise and Sunset are zero when the service reports none (polar day
	// and night).
	Sunrise, Sunset time.Time
}

// Forecast combines current conditions and the daily outlook.
type Forecast struct {
	Location  Location
	Timezone  string
	Current   Current
	Days      []Day
	FetchedAt time.Time
}

// IsDayAt tells whether the sun is up at the place at the given moment. It
// follows the sunrise and sunset of the forecast days, so the answer stays
// right between two fetches; the service's own day flag from the time of the
// fetch is the fallback outside the known sun times.
func (f Forecast) IsDayAt(now time.Time) bool {
	var last, first time.Time
	day, firstIsSunrise := f.Current.IsDay, false
	for _, d := range f.Days {
		for _, e := range []struct {
			at   time.Time
			rise bool
		}{{d.Sunrise, true}, {d.Sunset, false}} {
			if e.at.IsZero() {
				continue
			}
			if first.IsZero() || e.at.Before(first) {
				first, firstIsSunrise = e.at, e.rise
			}
			if !e.at.After(now) && e.at.After(last) {
				last, day = e.at, e.rise
			}
		}
	}
	switch {
	case first.IsZero():
		return f.Current.IsDay
	case last.IsZero():
		// Before the first sun event: the night before a sunrise, the day before a sunset.
		return !firstIsSunrise
	case now.Sub(last) > 24*time.Hour:
		// Past the last forecast day.
		return f.Current.IsDay
	}
	return day
}

// Place is a geocoding search result.
type Place struct {
	Name    string
	Region  string
	Country string
	Lat     float64
	Lon     float64
}

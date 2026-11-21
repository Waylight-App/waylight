package domain

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// MaxNameLength limits the name of a place ("Home", "Office").
const MaxNameLength = 80

// MaxLabelLength limits the address line shown under the name.
const MaxLabelLength = 200

// MaxPlaces bounds the places one account keeps; a picker with more is no
// longer a quick choice.
const MaxPlaces = 50

// ErrNotFound is returned for unknown places.
var ErrNotFound = kernel.NewPublicError(kernel.ErrNotFound, "The place does not exist.")

// ErrTooMany is returned when an account already keeps MaxPlaces.
var ErrTooMany = kernel.NewPublicErrorf(kernel.ErrInvalid, "At most {max} places can be saved.", "max", MaxPlaces)

// Place is a location entered once and chosen by widgets (#591): weather,
// fuel prices and departures take their coordinates from it.
type Place struct {
	ID string
	// Name is what the person calls it, e.g. "Home".
	Name string
	// Label is the address or place name the search found, e.g.
	// "Oldenburg, Lower Saxony, Germany"; empty for a browser location.
	Label     string
	Lat, Lon  float64
	CreatedAt time.Time
}

// NewPlace validates and creates a place.
func NewPlace(id, name, label string, lat, lon float64, now time.Time) (Place, error) {
	p := Place{ID: id, CreatedAt: now}
	if err := p.Change(name, label, lat, lon); err != nil {
		return Place{}, err
	}
	return p, nil
}

// Change renames or moves the place after the same checks as NewPlace.
func (p *Place) Change(name, label string, lat, lon float64) error {
	name, label = strings.TrimSpace(name), strings.TrimSpace(label)
	if name == "" || utf8.RuneCountInString(name) > MaxNameLength {
		return kernel.NewValidationError("name", "The name of a place must be between 1 and 80 characters long.")
	}
	if utf8.RuneCountInString(label) > MaxLabelLength {
		return kernel.NewValidationError("label", "The address of a place must be at most 200 characters long.")
	}
	if math.IsNaN(lat) || math.IsNaN(lon) || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return kernel.NewValidationError("lat", "The coordinates of the place are not valid.")
	}
	p.Name, p.Label, p.Lat, p.Lon = name, label, lat, lon
	return nil
}

// Use is a widget that takes its location from a place.
type Use struct {
	WidgetID, TabID, Kind string
}

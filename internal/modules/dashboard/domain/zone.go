package domain

import "waylight/internal/shared/kernel"

// Zone is the part of a zoned tab a widget sits in (#619). A tab with zones
// shows Today, My things and System one below the other instead of the free
// grid. The empty zone lets the client pick one by the widget's kind, so a
// widget placed before zones existed lands in a sensible place.
type Zone string

// The zones in their display order.
const (
	ZoneAuto   Zone = ""
	ZoneToday  Zone = "today"
	ZoneMine   Zone = "mine"
	ZoneSystem Zone = "system"
)

// Valid reports whether z is one of the zones or empty.
func (z Zone) Valid() bool {
	switch z {
	case ZoneAuto, ZoneToday, ZoneMine, ZoneSystem:
		return true
	}
	return false
}

// errInvalidZone is returned for a zone this version does not know.
func errInvalidZone() error {
	return kernel.NewValidationError("zone", "The zone must be today, mine or system.")
}

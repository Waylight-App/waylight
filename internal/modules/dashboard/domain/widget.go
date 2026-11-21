package domain

import (
	"regexp"

	"waylight/internal/shared/kernel"
)

// MaxConfigBytes limits the opaque widget configuration.
const MaxConfigBytes = 16 * 1024

var kindPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

// Widget is a tile on a tab. Its configuration is an opaque JSON document
// owned by the module that renders the widget.
type Widget struct {
	id     string
	tabID  string
	kind   string
	rect   Rect
	config string
	// size decides the web view (#595); empty for a freely sized tile.
	size Size
	// zone places the widget on a zoned tab (#619); empty = by kind.
	zone Zone
}

// NewWidget creates a widget after validating kind and configuration.
func NewWidget(id, tabID, kind string, rect Rect, config string) (Widget, error) {
	if !kindPattern.MatchString(kind) {
		return Widget{}, kernel.NewValidationError("kind", "The widget type is invalid.")
	}
	w := Widget{id: id, tabID: tabID, kind: kind, rect: rect}
	if err := w.Configure(config); err != nil {
		return Widget{}, err
	}
	return w, nil
}

// RestoreWidget rebuilds a widget from persisted state. A size this
// version does not know falls back to a freely sized tile.
func RestoreWidget(id, tabID, kind string, rect Rect, config string, size Size) Widget {
	if !size.Valid() {
		size = SizeAuto
	}
	return Widget{id: id, tabID: tabID, kind: kind, rect: rect, config: config, size: size}
}

// ID returns the identifier.
func (w Widget) ID() string { return w.id }

// TabID returns the tab the widget belongs to.
func (w Widget) TabID() string { return w.tabID }

// Kind returns the widget type, e.g. "weather".
func (w Widget) Kind() string { return w.kind }

// Rect returns position and size.
func (w Widget) Rect() Rect { return w.rect }

// Config returns the JSON configuration.
func (w Widget) Config() string { return w.config }

// Configure replaces the configuration. An empty configuration becomes "{}".
func (w *Widget) Configure(config string) error {
	if config == "" {
		config = "{}"
	}
	if len(config) > MaxConfigBytes {
		return kernel.NewValidationError("config", "The widget configuration is too large.")
	}
	w.config = config
	return nil
}

// Size returns the widget size on the web grid; empty for a freely sized tile.
func (w Widget) Size() Size { return w.size }

// Place moves and resizes the widget.
func (w *Widget) Place(r Rect) { w.rect = r }

// Resize sets the widget size (#595). The rectangle is placed separately,
// because the client fits the size into the grid around it.
func (w *Widget) Resize(s Size) error {
	if !s.Valid() {
		return kernel.NewValidationError("size", "The size must be small, medium, large or strip.")
	}
	w.size = s
	return nil
}

// Zone returns the zone the widget was moved to; empty while the client
// places it by kind.
func (w Widget) Zone() Zone { return w.zone }

// MoveToZone moves the widget to a zone of a zoned tab (#619); empty hands
// the choice back to the client's default by kind.
func (w *Widget) MoveToZone(z Zone) error {
	if !z.Valid() {
		return errInvalidZone()
	}
	w.zone = z
	return nil
}

// WithZone returns the widget restored with a stored zone. A zone this
// version does not know falls back to the default by kind.
func (w Widget) WithZone(z Zone) Widget {
	if !z.Valid() {
		z = ZoneAuto
	}
	w.zone = z
	return w
}

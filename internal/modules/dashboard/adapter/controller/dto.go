package controller

import "encoding/json"

// TabNameDTO is the body of POST /api/dashboard/tabs and PATCH /api/dashboard/tabs/{id}.
type TabNameDTO struct {
	Name string `json:"name"`
}

// TabOrderDTO is the body of PUT /api/dashboard/tabs/order.
type TabOrderDTO struct {
	TabIDs []string `json:"tabIds"`
}

// AddWidgetDTO is the body of POST /api/dashboard/tabs/{id}/widgets.
type AddWidgetDTO struct {
	Kind   string          `json:"kind"`
	X      int             `json:"x"`
	Y      int             `json:"y"`
	W      int             `json:"w"`
	H      int             `json:"h"`
	Config json.RawMessage `json:"config,omitempty" tstype:"{ [key: string]: unknown }"`
	// Size is small, medium, large or strip (#595); empty for a freely sized tile.
	Size string `json:"size,omitempty"`
}

// WidgetConfigDTO is the body of PATCH /api/dashboard/widgets/{id}.
type WidgetConfigDTO struct {
	Config json.RawMessage `json:"config" tstype:"{ [key: string]: unknown }"`
}

// PlacementDTO is the rectangle of one widget.
type PlacementDTO struct {
	ID string `json:"id"`
	X  int    `json:"x"`
	Y  int    `json:"y"`
	W  int    `json:"w"`
	H  int    `json:"h"`
	// Size sets the widget size (#595): small, medium, large, strip, or
	// empty for a freely sized tile. Omitted keeps the stored size.
	Size *string `json:"size,omitempty"`
	// Zone moves the widget to a zone of a zoned tab (#619): today, mine,
	// system, or empty for the default by kind. Omitted keeps the stored zone.
	Zone *string `json:"zone,omitempty"`
}

// LayoutDTO is the body of PUT /api/dashboard/tabs/{id}/layout.
type LayoutDTO struct {
	Items []PlacementDTO `json:"items"`
}

// TabBackgroundDTO is the body of PUT /api/dashboard/tabs/{id}/background;
// an empty background uses the global one.
type TabBackgroundDTO struct {
	Background string `json:"background"`
}

// TabZonesDTO is the body of PUT /api/dashboard/tabs/{id}/zones (#619).
type TabZonesDTO struct {
	Enabled bool `json:"enabled"`
}

// TabRuleDTO is one default-tab rule (#71).
type TabRuleDTO struct {
	Days     []int    `json:"days"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Networks []string `json:"networks"`
}

// TabRulesDTO is the body of PUT /api/dashboard/tabs/{id}/rules.
type TabRulesDTO struct {
	Rules []TabRuleDTO `json:"rules"`
}

// MobileItemInputDTO places a widget in the phone layout (#92).
type MobileItemInputDTO struct {
	ID     string `json:"id"`
	Hidden bool   `json:"hidden"`
	Rows   int    `json:"rows"`
	// Size is small, medium, large or strip; empty keeps full width (#585).
	Size string `json:"size,omitempty"`
}

// MobileLayoutDTO is the body of PUT /api/dashboard/tabs/{id}/mobile-layout.
type MobileLayoutDTO struct {
	Items []MobileItemInputDTO `json:"items"`
}

// SaveTemplateDTO is the body of POST /api/dashboard/templates.
type SaveTemplateDTO struct {
	TabID string `json:"tabId"`
	Name  string `json:"name"`
}

// BuiltinTabDTO is the body of POST /api/dashboard/templates/builtin/{id}/tabs.
type BuiltinTabDTO struct {
	// Name of the new tab; empty uses the template's.
	Name string `json:"name"`
	// Location fills the weather widgets; omitted leaves them to be set up.
	Location *LocationDTO `json:"location,omitempty"`
}

// LocationDTO is a place for the weather widgets of a template.
type LocationDTO struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Label   string  `json:"label"`
	PlaceID string  `json:"placeId"`
}

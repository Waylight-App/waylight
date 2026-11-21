// Package presenter maps places to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"waylight/internal/modules/places/usecase"
)

// UseDTO is a widget that takes its location from a place.
type UseDTO struct {
	WidgetID string `json:"widgetId"`
	TabID    string `json:"tabId"`
	Kind     string `json:"kind"`
}

// PlaceDTO is one place in API responses.
type PlaceDTO struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Label string  `json:"label"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	// UsedBy lists the widgets on the account's own tabs that use the place.
	UsedBy []UseDTO `json:"usedBy"`
}

// Place maps one place.
func Place(p usecase.PlaceView) PlaceDTO {
	out := PlaceDTO{ID: p.ID, Name: p.Name, Label: p.Label, Lat: p.Lat, Lon: p.Lon, UsedBy: make([]UseDTO, 0, len(p.UsedBy))}
	for _, u := range p.UsedBy {
		out.UsedBy = append(out.UsedBy, UseDTO{WidgetID: u.WidgetID, TabID: u.TabID, Kind: u.Kind})
	}
	return out
}

// Places maps a list of places.
func Places(list []usecase.PlaceView) []PlaceDTO {
	out := make([]PlaceDTO, 0, len(list))
	for _, p := range list {
		out = append(out, Place(p))
	}
	return out
}

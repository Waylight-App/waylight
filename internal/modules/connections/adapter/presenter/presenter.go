// Package presenter maps connections to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import "waylight/internal/modules/connections/usecase"

// UseDTO is a widget whose settings name a connection.
type UseDTO struct {
	WidgetID string `json:"widgetId"`
	TabID    string `json:"tabId"`
	Kind     string `json:"kind"`
}

// ConnectionDTO is one connection in API responses. It carries the names
// of the credential fields that are set, never their values.
type ConnectionDTO struct {
	ID      string `json:"id"`
	Service string `json:"service"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	// Fields lists the credential fields that are set, e.g. ["apiKey"].
	Fields []string `json:"fields"`
	UsedBy []UseDTO `json:"usedBy"`
}

// Connection maps one connection.
func Connection(v usecase.View) ConnectionDTO {
	out := ConnectionDTO{ID: v.ID, Service: v.Service, Name: v.Name, URL: v.URL,
		Fields: append([]string{}, v.Fields...), UsedBy: make([]UseDTO, 0, len(v.UsedBy))}
	for _, u := range v.UsedBy {
		out.UsedBy = append(out.UsedBy, UseDTO{WidgetID: u.WidgetID, TabID: u.TabID, Kind: u.Kind})
	}
	return out
}

// Connections maps a list.
func Connections(list []usecase.View) []ConnectionDTO {
	out := make([]ConnectionDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Connection(v))
	}
	return out
}

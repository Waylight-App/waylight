// Package presenter maps connections to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"maps"

	"waylight/internal/modules/integrations/domain"
	"waylight/internal/modules/integrations/usecase"
)

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
	// Values holds the fields that are not secret, e.g. the user name.
	Values map[string]string `json:"values"`
	UsedBy []UseDTO          `json:"usedBy"`
}

// Connection maps one connection.
func Connection(v usecase.View) ConnectionDTO {
	out := ConnectionDTO{ID: v.ID, Service: v.Service, Name: v.Name, URL: v.URL,
		Fields: append([]string{}, v.Fields...), Values: map[string]string{}, UsedBy: make([]UseDTO, 0, len(v.UsedBy))}
	maps.Copy(out.Values, v.Values)
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

// FieldDTO is one credential field of a service.
type FieldDTO struct {
	Name     string `json:"name"`
	Secret   bool   `json:"secret"`
	Optional bool   `json:"optional"`
}

// ServiceDTO is a service a connection can be made to.
type ServiceDTO struct {
	Service string     `json:"service"`
	Title   string     `json:"title"`
	Fields  []FieldDTO `json:"fields"`
	// FixedURL is set for services with a single address.
	FixedURL string `json:"fixedUrl,omitempty"`
}

// Services maps the specs.
func Services(list []domain.Spec) []ServiceDTO {
	out := make([]ServiceDTO, 0, len(list))
	for _, s := range list {
		d := ServiceDTO{Service: s.Service, Title: s.Title, FixedURL: s.FixedURL, Fields: make([]FieldDTO, 0, len(s.Fields))}
		for _, f := range s.Fields {
			d.Fields = append(d.Fields, FieldDTO{Name: f.Name, Secret: f.Secret, Optional: f.Optional})
		}
		out = append(out, d)
	}
	return out
}

// Package presenter maps the attention list to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/attention/domain"
	"waylight/internal/shared/kernel"
)

// ReasonDTO is a translatable message: clients look up MessageKey and fill
// in Params; Message is the English text for clients without it (#102).
type ReasonDTO struct {
	Message    string            `json:"message"`
	MessageKey string            `json:"messageKey"`
	Params     map[string]string `json:"params,omitempty"`
}

// ItemDTO is one entry of GET /api/attention.
type ItemDTO struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	Integration  string    `json:"integration"`
	WidgetID     string    `json:"widgetId,omitempty"`
	TabID        string    `json:"tabId,omitempty"`
	ConnectionID string    `json:"connectionId,omitempty"`
	Title        string    `json:"title"`
	Reason       ReasonDTO `json:"reason"`
	Since        time.Time `json:"since"`
	Action       string    `json:"action"`
	Field        string    `json:"field,omitempty"`
}

// ListDTO is the answer of GET /api/attention.
type ListDTO struct {
	Items []ItemDTO `json:"items"`
	Count int       `json:"count"`
}

// List maps the sorted entries to the DTO; an empty list stays an empty
// array, never null.
func List(items []domain.Item) ListDTO {
	out := ListDTO{Items: make([]ItemDTO, 0, len(items)), Count: len(items)}
	for _, it := range items {
		out.Items = append(out.Items, ItemDTO{
			ID: it.ID, Kind: string(it.Kind), Integration: it.Integration, WidgetID: it.WidgetID,
			TabID: it.TabID, ConnectionID: it.ConnectionID, Title: it.Title, Reason: reason(it.Reason),
			Since: it.Since.UTC(), Action: string(it.Action), Field: it.Field,
		})
	}
	return out
}

func reason(m domain.Message) ReasonDTO {
	text := m.Text
	if text == "" {
		text = m.Template
	}
	return ReasonDTO{Message: text, MessageKey: kernel.MessageKey(m.Template), Params: m.Params}
}

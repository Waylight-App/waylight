// Package presenter maps board notes to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"time"

	"waylight/internal/modules/board/usecase"
)

// NoteDTO is one note in API responses.
type NoteDTO struct {
	ID     string `json:"id"`
	Text   string `json:"text"`
	Pinned bool   `json:"pinned"`
	// ExpiresOn is the last day the note is shown, YYYY-MM-DD, or empty.
	ExpiresOn string `json:"expiresOn"`
	// Visibility is "private", "everyone" or "group:<id>" (ADR 0018).
	Visibility string `json:"visibility"`
	// Mine is true when the viewer wrote the note and may change it.
	Mine      bool      `json:"mine"`
	CreatedAt time.Time `json:"createdAt"`
}

// Note maps one note.
func Note(v usecase.NoteView) NoteDTO {
	return NoteDTO{ID: v.ID, Text: v.Text, Pinned: v.Pinned, ExpiresOn: v.ExpiryDate(), Visibility: string(v.Access.Visibility), Mine: v.Mine, CreatedAt: v.CreatedAt}
}

// Notes maps a list of notes.
func Notes(list []usecase.NoteView) []NoteDTO {
	out := make([]NoteDTO, 0, len(list))
	for _, v := range list {
		out = append(out, Note(v))
	}
	return out
}

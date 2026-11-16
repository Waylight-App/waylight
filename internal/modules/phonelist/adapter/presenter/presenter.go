// Package presenter maps phone list entries to JSON DTOs.
//
// Ring: Interface Adapters.
package presenter

import (
	"trisse-home/internal/modules/phonelist/domain"
)

// EntryDTO is one person of the phone list.
type EntryDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Role  string `json:"role"`
	Phone string `json:"phone"`
	// Dial is the number for a tel: link (digits and a leading +); empty
	// without a phone number.
	Dial  string `json:"dial"`
	Email string `json:"email"`
	Team  string `json:"team"`
}

// Entry maps one entry.
func Entry(e domain.Entry) EntryDTO {
	return EntryDTO{ID: e.ID, Name: e.Name, Role: e.Role, Phone: e.Phone, Dial: domain.DialNumber(e.Phone), Email: e.Email, Team: e.Team}
}

// Entries maps a list of entries.
func Entries(list []domain.Entry) []EntryDTO {
	out := make([]EntryDTO, 0, len(list))
	for _, e := range list {
		out = append(out, Entry(e))
	}
	return out
}
